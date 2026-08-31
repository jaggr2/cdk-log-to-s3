// Command log-to-s3 is an AWS Lambda external extension. It subscribes to the
// Lambda Telemetry API, turns the delivered records into a columnar schema and
// writes them to S3 as Parquet, where Athena can query them directly.
//
// It is packaged as a layer whose only content is extensions/log-to-s3, so the
// runtime starts it automatically. See the @jaggr2/cdk-log-to-s3 constructs for
// the infrastructure side.
package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	cfg, warnings := loadConfig()
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "LOG_TO_S3_CONFIG_WARNING: %s\n", w)
	}

	// The bucket is the one setting with no useful default. Without it there
	// is nothing to subscribe for, but the extension must still register or
	// the function will not leave INIT.
	var up *Uploader
	if cfg.Bucket == "" {
		fmt.Fprintf(os.Stderr,
			"LOG_TO_S3_DISABLED: neither %s nor %s is set; logs stay in CloudWatch only\n",
			envBucket, envBucketDeprecated)
	} else {
		var err error
		up, err = NewUploader(context.Background(), cfg.Bucket, cfg.Prefix, cfg.Region)
		if err != nil {
			fmt.Fprintf(os.Stderr, "LOG_TO_S3_DISABLED: cannot build S3 client: %v\n", err)
		}
	}

	listener := NewListener(cfg, up)
	if err := listener.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "LOG_TO_S3_DISABLED: cannot start telemetry listener: %v\n", err)
	}

	extensionID, err := register(cfg.ExtensionName)
	if err != nil {
		// Fatal by design - see the comment on register().
		fmt.Fprintf(os.Stderr,
			"LOG_TO_S3_FATAL: could not register with the Extensions API: %v\n"+
				"Check that the layer contains extensions/%s and that the file is executable.\n",
			err, cfg.ExtensionName)
		os.Exit(1)
	}

	// Subscribing with no destination for the data would be pure overhead.
	subscribed := false
	if up != nil {
		if err := subscribeTelemetry(extensionID, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "LOG_TO_S3_ERROR: telemetry subscribe failed, S3 logging is off for this sandbox: %v\n", err)
		} else {
			subscribed = true
		}
	}

	// Event loop. Telemetry for invocation N is delivered after its INVOKE
	// event, so the buffer is flushed on the way into the next wait rather
	// than on the way out of the previous one.
	for {
		if subscribed && listener.Buffer().SizeBytes() > 0 {
			listener.flushLogged("invoke")
		}

		eventType, err := nextEvent(extensionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "LOG_TO_S3_ERROR: event loop ended: %v\n", err)
			break
		}
		if eventType == "SHUTDOWN" {
			break
		}
	}

	listener.Stop()
}

# API Reference <a name="API Reference" id="api-reference"></a>

## Constructs <a name="Constructs" id="Constructs"></a>

### LogAnalytics <a name="LogAnalytics" id="@jaggr2/cdk-log-to-s3.LogAnalytics"></a>

A Glue table and Athena workgroup for querying the Parquet files the extension writes.

Partitions are resolved with Athena partition projection, computed from the
key layout at query time. Nothing has to register partitions, so there is no
crawler, no scheduled MSCK REPAIR and no broad glue:* permission anywhere.

Because projection derives the S3 location from a template, every directory
level below the prefix must be a partition key - which is exactly the layout
the extension writes.

*Example*

```typescript
const analytics = LogAnalytics.fromExtension(this, 'Analytics', ext, {
  databaseName: 'my_app_logs',
});
```


#### Initializers <a name="Initializers" id="@jaggr2/cdk-log-to-s3.LogAnalytics.Initializer"></a>

```typescript
import { LogAnalytics } from '@jaggr2/cdk-log-to-s3'

new LogAnalytics(scope: Construct, id: string, props: LogAnalyticsProps)
```

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalytics.Initializer.parameter.scope">scope</a></code> | <code>constructs.Construct</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalytics.Initializer.parameter.id">id</a></code> | <code>string</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalytics.Initializer.parameter.props">props</a></code> | <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsProps">LogAnalyticsProps</a></code> | *No description.* |

---

##### `scope`<sup>Required</sup> <a name="scope" id="@jaggr2/cdk-log-to-s3.LogAnalytics.Initializer.parameter.scope"></a>

- *Type:* constructs.Construct

---

##### `id`<sup>Required</sup> <a name="id" id="@jaggr2/cdk-log-to-s3.LogAnalytics.Initializer.parameter.id"></a>

- *Type:* string

---

##### `props`<sup>Required</sup> <a name="props" id="@jaggr2/cdk-log-to-s3.LogAnalytics.Initializer.parameter.props"></a>

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsProps">LogAnalyticsProps</a>

---

#### Methods <a name="Methods" id="Methods"></a>

| **Name** | **Description** |
| --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalytics.toString">toString</a></code> | Returns a string representation of this construct. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalytics.with">with</a></code> | Applies one or more mixins to this construct. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalytics.grantQuery">grantQuery</a></code> | Grant everything needed to run a query against this table: the workgroup, the catalog entries, read access to the log prefix and read/write on the results bucket. |

---

##### `toString` <a name="toString" id="@jaggr2/cdk-log-to-s3.LogAnalytics.toString"></a>

```typescript
public toString(): string
```

Returns a string representation of this construct.

##### `with` <a name="with" id="@jaggr2/cdk-log-to-s3.LogAnalytics.with"></a>

```typescript
public with(mixins: ...IMixin[]): IConstruct
```

Applies one or more mixins to this construct.

Mixins are applied in order. The list of constructs is captured at the
start of the call, so constructs added by a mixin will not be visited.
Use multiple `with()` calls if subsequent mixins should apply to added
constructs.

###### `mixins`<sup>Required</sup> <a name="mixins" id="@jaggr2/cdk-log-to-s3.LogAnalytics.with.parameter.mixins"></a>

- *Type:* ...constructs.IMixin[]

The mixins to apply.

---

##### `grantQuery` <a name="grantQuery" id="@jaggr2/cdk-log-to-s3.LogAnalytics.grantQuery"></a>

```typescript
public grantQuery(grantee: IGrantable): void
```

Grant everything needed to run a query against this table: the workgroup, the catalog entries, read access to the log prefix and read/write on the results bucket.

Every statement is scoped to a resource.

###### `grantee`<sup>Required</sup> <a name="grantee" id="@jaggr2/cdk-log-to-s3.LogAnalytics.grantQuery.parameter.grantee"></a>

- *Type:* aws-cdk-lib.aws_iam.IGrantable

---

#### Static Functions <a name="Static Functions" id="Static Functions"></a>

| **Name** | **Description** |
| --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalytics.isConstruct">isConstruct</a></code> | Checks if `x` is a construct. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalytics.fromExtension">fromExtension</a></code> | Build the analytics stack for an extension, taking the bucket and the key prefix from it so the two cannot drift apart. |

---

##### `isConstruct` <a name="isConstruct" id="@jaggr2/cdk-log-to-s3.LogAnalytics.isConstruct"></a>

```typescript
import { LogAnalytics } from '@jaggr2/cdk-log-to-s3'

LogAnalytics.isConstruct(x: any)
```

Checks if `x` is a construct.

Use this method instead of `instanceof` to properly detect `Construct`
instances, even when the construct library is symlinked.

Explanation: in JavaScript, multiple copies of the `constructs` library on
disk are seen as independent, completely different libraries. As a
consequence, the class `Construct` in each copy of the `constructs` library
is seen as a different class, and an instance of one class will not test as
`instanceof` the other class. `npm install` will not create installations
like this, but users may manually symlink construct libraries together or
use a monorepo tool: in those cases, multiple copies of the `constructs`
library can be accidentally installed, and `instanceof` will behave
unpredictably. It is safest to avoid using `instanceof`, and using
this type-testing method instead.

###### `x`<sup>Required</sup> <a name="x" id="@jaggr2/cdk-log-to-s3.LogAnalytics.isConstruct.parameter.x"></a>

- *Type:* any

Any object.

---

##### `fromExtension` <a name="fromExtension" id="@jaggr2/cdk-log-to-s3.LogAnalytics.fromExtension"></a>

```typescript
import { LogAnalytics } from '@jaggr2/cdk-log-to-s3'

LogAnalytics.fromExtension(scope: Construct, id: string, extension: ILogToS3Extension, options: LogAnalyticsFromExtensionOptions)
```

Build the analytics stack for an extension, taking the bucket and the key prefix from it so the two cannot drift apart.

###### `scope`<sup>Required</sup> <a name="scope" id="@jaggr2/cdk-log-to-s3.LogAnalytics.fromExtension.parameter.scope"></a>

- *Type:* constructs.Construct

---

###### `id`<sup>Required</sup> <a name="id" id="@jaggr2/cdk-log-to-s3.LogAnalytics.fromExtension.parameter.id"></a>

- *Type:* string

---

###### `extension`<sup>Required</sup> <a name="extension" id="@jaggr2/cdk-log-to-s3.LogAnalytics.fromExtension.parameter.extension"></a>

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.ILogToS3Extension">ILogToS3Extension</a>

---

###### `options`<sup>Required</sup> <a name="options" id="@jaggr2/cdk-log-to-s3.LogAnalytics.fromExtension.parameter.options"></a>

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions">LogAnalyticsFromExtensionOptions</a>

---

#### Properties <a name="Properties" id="Properties"></a>

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalytics.property.node">node</a></code> | <code>constructs.Node</code> | The tree node. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalytics.property.databaseName">databaseName</a></code> | <code>string</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalytics.property.resultsBucket">resultsBucket</a></code> | <code>aws-cdk-lib.aws_s3.IBucket</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalytics.property.table">table</a></code> | <code>aws-cdk-lib.aws_glue.CfnTable</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalytics.property.tableName">tableName</a></code> | <code>string</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalytics.property.database">database</a></code> | <code>aws-cdk-lib.aws_glue.CfnDatabase</code> | The created database, or undefined when createDatabase was false. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalytics.property.workgroup">workgroup</a></code> | <code>aws-cdk-lib.aws_athena.CfnWorkGroup</code> | The created workgroup, or undefined when createWorkgroup was false. |

---

##### `node`<sup>Required</sup> <a name="node" id="@jaggr2/cdk-log-to-s3.LogAnalytics.property.node"></a>

```typescript
public readonly node: Node;
```

- *Type:* constructs.Node

The tree node.

---

##### `databaseName`<sup>Required</sup> <a name="databaseName" id="@jaggr2/cdk-log-to-s3.LogAnalytics.property.databaseName"></a>

```typescript
public readonly databaseName: string;
```

- *Type:* string

---

##### `resultsBucket`<sup>Required</sup> <a name="resultsBucket" id="@jaggr2/cdk-log-to-s3.LogAnalytics.property.resultsBucket"></a>

```typescript
public readonly resultsBucket: IBucket;
```

- *Type:* aws-cdk-lib.aws_s3.IBucket

---

##### `table`<sup>Required</sup> <a name="table" id="@jaggr2/cdk-log-to-s3.LogAnalytics.property.table"></a>

```typescript
public readonly table: CfnTable;
```

- *Type:* aws-cdk-lib.aws_glue.CfnTable

---

##### `tableName`<sup>Required</sup> <a name="tableName" id="@jaggr2/cdk-log-to-s3.LogAnalytics.property.tableName"></a>

```typescript
public readonly tableName: string;
```

- *Type:* string

---

##### `database`<sup>Optional</sup> <a name="database" id="@jaggr2/cdk-log-to-s3.LogAnalytics.property.database"></a>

```typescript
public readonly database: CfnDatabase;
```

- *Type:* aws-cdk-lib.aws_glue.CfnDatabase

The created database, or undefined when createDatabase was false.

---

##### `workgroup`<sup>Optional</sup> <a name="workgroup" id="@jaggr2/cdk-log-to-s3.LogAnalytics.property.workgroup"></a>

```typescript
public readonly workgroup: CfnWorkGroup;
```

- *Type:* aws-cdk-lib.aws_athena.CfnWorkGroup

The created workgroup, or undefined when createWorkgroup was false.

---


### LogBucket <a name="LogBucket" id="@jaggr2/cdk-log-to-s3.LogBucket"></a>

An S3 bucket with defaults suited to Parquet log storage: private, encrypted, TLS-only, and on a tiered lifecycle.

It extends s3.Bucket, so it can be passed anywhere an IBucket is accepted -
including LogToS3ExtensionProps.logsBucket and LogAnalyticsProps.logsBucket.

The removal policy is RETAIN. Logs usually outlive the stack that produced
them, and an accidental `cdk destroy` should not be able to delete an audit
trail. Pass `removalPolicy: RemovalPolicy.DESTROY` for throwaway stacks.

#### Initializers <a name="Initializers" id="@jaggr2/cdk-log-to-s3.LogBucket.Initializer"></a>

```typescript
import { LogBucket } from '@jaggr2/cdk-log-to-s3'

new LogBucket(scope: Construct, id: string, props?: LogBucketProps)
```

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.Initializer.parameter.scope">scope</a></code> | <code>constructs.Construct</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.Initializer.parameter.id">id</a></code> | <code>string</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.Initializer.parameter.props">props</a></code> | <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps">LogBucketProps</a></code> | *No description.* |

---

##### `scope`<sup>Required</sup> <a name="scope" id="@jaggr2/cdk-log-to-s3.LogBucket.Initializer.parameter.scope"></a>

- *Type:* constructs.Construct

---

##### `id`<sup>Required</sup> <a name="id" id="@jaggr2/cdk-log-to-s3.LogBucket.Initializer.parameter.id"></a>

- *Type:* string

---

##### `props`<sup>Optional</sup> <a name="props" id="@jaggr2/cdk-log-to-s3.LogBucket.Initializer.parameter.props"></a>

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogBucketProps">LogBucketProps</a>

---

#### Methods <a name="Methods" id="Methods"></a>

| **Name** | **Description** |
| --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.toString">toString</a></code> | Returns a string representation of this construct. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.with">with</a></code> | Applies one or more mixins to this construct. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.applyRemovalPolicy">applyRemovalPolicy</a></code> | Apply the given removal policy to this resource. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.addEventNotification">addEventNotification</a></code> | Adds a bucket notification event destination. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.addObjectCreatedNotification">addObjectCreatedNotification</a></code> | Subscribes a destination to receive notifications when an object is created in the bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.addObjectRemovedNotification">addObjectRemovedNotification</a></code> | Subscribes a destination to receive notifications when an object is removed from the bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.addToResourcePolicy">addToResourcePolicy</a></code> | Adds a statement to the resource policy for a principal (i.e. account/role/service) to perform actions on this bucket and/or its contents. Use `bucketArn` and `arnForObjects(keys)` to obtain ARNs for this bucket or objects. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.arnForObjects">arnForObjects</a></code> | Returns an ARN that represents all objects within the bucket that match the key pattern specified. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.enableEventBridgeNotification">enableEventBridgeNotification</a></code> | Enables event bridge notification, causing all events below to be sent to EventBridge:. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.grantDelete">grantDelete</a></code> | Grants s3:DeleteObject* permission to an IAM principal for objects in this bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.grantPublicAccess">grantPublicAccess</a></code> | Allows unrestricted access to objects from this bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.grantPut">grantPut</a></code> | Grants s3:PutObject* and s3:Abort* permissions for this bucket to an IAM principal. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.grantPutAcl">grantPutAcl</a></code> | Grant the given IAM identity permissions to modify the ACLs of objects in the given Bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.grantRead">grantRead</a></code> | Grant read permissions for this bucket and it's contents to an IAM principal (Role/Group/User). |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.grantReadWrite">grantReadWrite</a></code> | Grants read/write permissions for this bucket and it's contents to an IAM principal (Role/Group/User). |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.grantWrite">grantWrite</a></code> | Grant write permissions to this bucket to an IAM principal. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.onCloudTrailEvent">onCloudTrailEvent</a></code> | Define a CloudWatch event that triggers when something happens to this repository. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.onCloudTrailPutObject">onCloudTrailPutObject</a></code> | Defines an AWS CloudWatch event that triggers when an object is uploaded to the specified paths (keys) in this bucket using the PutObject API call. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.onCloudTrailWriteObject">onCloudTrailWriteObject</a></code> | Defines an AWS CloudWatch event that triggers when an object at the specified paths (keys) in this bucket are written to. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.s3UrlForObject">s3UrlForObject</a></code> | The S3 URL of an S3 object. For example:. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.transferAccelerationUrlForObject">transferAccelerationUrlForObject</a></code> | The https Transfer Acceleration URL of an S3 object. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.urlForObject">urlForObject</a></code> | The https URL of an S3 object. Specify `regional: false` at the options for non-regional URLs. For example:. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.virtualHostedUrlForObject">virtualHostedUrlForObject</a></code> | The virtual hosted-style URL of an S3 object. Specify `regional: false` at the options for non-regional URL. For example:. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.addCorsRule">addCorsRule</a></code> | Adds a cross-origin access configuration for objects in an Amazon S3 bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.addInventory">addInventory</a></code> | Add an inventory configuration. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.addLifecycleRule">addLifecycleRule</a></code> | Add a lifecycle rule to the bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.addMetric">addMetric</a></code> | Adds a metrics configuration for the CloudWatch request metrics from the bucket. |

---

##### `toString` <a name="toString" id="@jaggr2/cdk-log-to-s3.LogBucket.toString"></a>

```typescript
public toString(): string
```

Returns a string representation of this construct.

##### `with` <a name="with" id="@jaggr2/cdk-log-to-s3.LogBucket.with"></a>

```typescript
public with(mixins: ...IMixin[]): IConstruct
```

Applies one or more mixins to this construct.

Mixins are applied in order. The list of constructs is captured at the
start of the call, so constructs added by a mixin will not be visited.
Use multiple `with()` calls if subsequent mixins should apply to added
constructs.

###### `mixins`<sup>Required</sup> <a name="mixins" id="@jaggr2/cdk-log-to-s3.LogBucket.with.parameter.mixins"></a>

- *Type:* ...constructs.IMixin[]

The mixins to apply.

---

##### `applyRemovalPolicy` <a name="applyRemovalPolicy" id="@jaggr2/cdk-log-to-s3.LogBucket.applyRemovalPolicy"></a>

```typescript
public applyRemovalPolicy(policy: RemovalPolicy): void
```

Apply the given removal policy to this resource.

The Removal Policy controls what happens to this resource when it stops
being managed by CloudFormation, either because you've removed it from the
CDK application or because you've made a change that requires the resource
to be replaced.

The resource can be deleted (`RemovalPolicy.DESTROY`), or left in your AWS
account for data recovery and cleanup later (`RemovalPolicy.RETAIN`).

###### `policy`<sup>Required</sup> <a name="policy" id="@jaggr2/cdk-log-to-s3.LogBucket.applyRemovalPolicy.parameter.policy"></a>

- *Type:* aws-cdk-lib.RemovalPolicy

---

##### `addEventNotification` <a name="addEventNotification" id="@jaggr2/cdk-log-to-s3.LogBucket.addEventNotification"></a>

```typescript
public addEventNotification(event: EventType, dest: IBucketNotificationDestination, filters: ...NotificationKeyFilter[]): void
```

Adds a bucket notification event destination.

> [https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html](https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html)

*Example*

```typescript
   declare const myLambda: lambda.Function;
   const bucket = new s3.Bucket(this, 'MyBucket');
   bucket.addEventNotification(s3.EventType.OBJECT_CREATED, new s3n.LambdaDestination(myLambda), {prefix: 'home/myusername/*'});
```


###### `event`<sup>Required</sup> <a name="event" id="@jaggr2/cdk-log-to-s3.LogBucket.addEventNotification.parameter.event"></a>

- *Type:* aws-cdk-lib.aws_s3.EventType

The event to trigger the notification.

---

###### `dest`<sup>Required</sup> <a name="dest" id="@jaggr2/cdk-log-to-s3.LogBucket.addEventNotification.parameter.dest"></a>

- *Type:* aws-cdk-lib.aws_s3.IBucketNotificationDestination

The notification destination (Lambda, SNS Topic or SQS Queue).

---

###### `filters`<sup>Required</sup> <a name="filters" id="@jaggr2/cdk-log-to-s3.LogBucket.addEventNotification.parameter.filters"></a>

- *Type:* ...aws-cdk-lib.aws_s3.NotificationKeyFilter[]

S3 object key filter rules to determine which objects trigger this event.

Each filter must include a `prefix` and/or `suffix`
that will be matched against the s3 object key. Refer to the S3 Developer Guide
for details about allowed filter rules.

---

##### `addObjectCreatedNotification` <a name="addObjectCreatedNotification" id="@jaggr2/cdk-log-to-s3.LogBucket.addObjectCreatedNotification"></a>

```typescript
public addObjectCreatedNotification(dest: IBucketNotificationDestination, filters: ...NotificationKeyFilter[]): void
```

Subscribes a destination to receive notifications when an object is created in the bucket.

This is identical to calling
`onEvent(EventType.OBJECT_CREATED)`.

###### `dest`<sup>Required</sup> <a name="dest" id="@jaggr2/cdk-log-to-s3.LogBucket.addObjectCreatedNotification.parameter.dest"></a>

- *Type:* aws-cdk-lib.aws_s3.IBucketNotificationDestination

The notification destination (see onEvent).

---

###### `filters`<sup>Required</sup> <a name="filters" id="@jaggr2/cdk-log-to-s3.LogBucket.addObjectCreatedNotification.parameter.filters"></a>

- *Type:* ...aws-cdk-lib.aws_s3.NotificationKeyFilter[]

Filters (see onEvent).

---

##### `addObjectRemovedNotification` <a name="addObjectRemovedNotification" id="@jaggr2/cdk-log-to-s3.LogBucket.addObjectRemovedNotification"></a>

```typescript
public addObjectRemovedNotification(dest: IBucketNotificationDestination, filters: ...NotificationKeyFilter[]): void
```

Subscribes a destination to receive notifications when an object is removed from the bucket.

This is identical to calling
`onEvent(EventType.OBJECT_REMOVED)`.

###### `dest`<sup>Required</sup> <a name="dest" id="@jaggr2/cdk-log-to-s3.LogBucket.addObjectRemovedNotification.parameter.dest"></a>

- *Type:* aws-cdk-lib.aws_s3.IBucketNotificationDestination

The notification destination (see onEvent).

---

###### `filters`<sup>Required</sup> <a name="filters" id="@jaggr2/cdk-log-to-s3.LogBucket.addObjectRemovedNotification.parameter.filters"></a>

- *Type:* ...aws-cdk-lib.aws_s3.NotificationKeyFilter[]

Filters (see onEvent).

---

##### `addToResourcePolicy` <a name="addToResourcePolicy" id="@jaggr2/cdk-log-to-s3.LogBucket.addToResourcePolicy"></a>

```typescript
public addToResourcePolicy(permission: PolicyStatement): AddToResourcePolicyResult
```

Adds a statement to the resource policy for a principal (i.e. account/role/service) to perform actions on this bucket and/or its contents. Use `bucketArn` and `arnForObjects(keys)` to obtain ARNs for this bucket or objects.

Note that the policy statement may or may not be added to the policy.
For example, when an `IBucket` is created from an existing bucket,
it's not possible to tell whether the bucket already has a policy
attached, let alone to re-use that policy to add more statements to it.
So it's safest to do nothing in these cases.

###### `permission`<sup>Required</sup> <a name="permission" id="@jaggr2/cdk-log-to-s3.LogBucket.addToResourcePolicy.parameter.permission"></a>

- *Type:* aws-cdk-lib.aws_iam.PolicyStatement

the policy statement to be added to the bucket's policy.

---

##### `arnForObjects` <a name="arnForObjects" id="@jaggr2/cdk-log-to-s3.LogBucket.arnForObjects"></a>

```typescript
public arnForObjects(keyPattern: string): string
```

Returns an ARN that represents all objects within the bucket that match the key pattern specified.

To represent all keys, specify ``"*"``.

If you need to specify a keyPattern with multiple components, concatenate them into a single string, e.g.:

  arnForObjects(`home/${team}/${user}/*`)

###### `keyPattern`<sup>Required</sup> <a name="keyPattern" id="@jaggr2/cdk-log-to-s3.LogBucket.arnForObjects.parameter.keyPattern"></a>

- *Type:* string

---

##### `enableEventBridgeNotification` <a name="enableEventBridgeNotification" id="@jaggr2/cdk-log-to-s3.LogBucket.enableEventBridgeNotification"></a>

```typescript
public enableEventBridgeNotification(): void
```

Enables event bridge notification, causing all events below to be sent to EventBridge:.

Object Deleted (DeleteObject)
- Object Deleted (Lifecycle expiration)
- Object Restore Initiated
- Object Restore Completed
- Object Restore Expired
- Object Storage Class Changed
- Object Access Tier Changed
- Object ACL Updated
- Object Tags Added
- Object Tags Deleted

##### `grantDelete` <a name="grantDelete" id="@jaggr2/cdk-log-to-s3.LogBucket.grantDelete"></a>

```typescript
public grantDelete(identity: IGrantable, objectsKeyPattern?: any): Grant
```

Grants s3:DeleteObject* permission to an IAM principal for objects in this bucket.

###### `identity`<sup>Required</sup> <a name="identity" id="@jaggr2/cdk-log-to-s3.LogBucket.grantDelete.parameter.identity"></a>

- *Type:* aws-cdk-lib.aws_iam.IGrantable

The principal.

---

###### `objectsKeyPattern`<sup>Optional</sup> <a name="objectsKeyPattern" id="@jaggr2/cdk-log-to-s3.LogBucket.grantDelete.parameter.objectsKeyPattern"></a>

- *Type:* any

Restrict the permission to a certain key pattern (default '*').

Parameter type is `any` but `string` should be passed in.

---

##### `grantPublicAccess` <a name="grantPublicAccess" id="@jaggr2/cdk-log-to-s3.LogBucket.grantPublicAccess"></a>

```typescript
public grantPublicAccess(allowedActions: ...string[], keyPrefix?: string): Grant
```

Allows unrestricted access to objects from this bucket.

IMPORTANT: This permission allows anyone to perform actions on S3 objects
in this bucket, which is useful for when you configure your bucket as a
website and want everyone to be able to read objects in the bucket without
needing to authenticate.

Without arguments, this method will grant read ("s3:GetObject") access to
all objects ("*") in the bucket.

The method returns the `iam.Grant` object, which can then be modified
as needed. For example, you can add a condition that will restrict access only
to an IPv4 range like this:

    const grant = bucket.grantPublicAccess();
    grant.resourceStatement!.addCondition(‘IpAddress’, { “aws:SourceIp”: “54.240.143.0/24” });

Note that if this `IBucket` refers to an existing bucket, possibly not
managed by CloudFormation, this method will have no effect, since it's
impossible to modify the policy of an existing bucket.

###### `allowedActions`<sup>Required</sup> <a name="allowedActions" id="@jaggr2/cdk-log-to-s3.LogBucket.grantPublicAccess.parameter.allowedActions"></a>

- *Type:* ...string[]

the set of S3 actions to allow.

Default is "s3:GetObject".

---

###### `keyPrefix`<sup>Optional</sup> <a name="keyPrefix" id="@jaggr2/cdk-log-to-s3.LogBucket.grantPublicAccess.parameter.keyPrefix"></a>

- *Type:* string

the prefix of S3 object keys (e.g. `home/*`). Default is "*".

---

##### `grantPut` <a name="grantPut" id="@jaggr2/cdk-log-to-s3.LogBucket.grantPut"></a>

```typescript
public grantPut(identity: IGrantable, objectsKeyPattern?: any): Grant
```

Grants s3:PutObject* and s3:Abort* permissions for this bucket to an IAM principal.

If encryption is used, permission to use the key to encrypt the contents
of written files will also be granted to the same principal.

###### `identity`<sup>Required</sup> <a name="identity" id="@jaggr2/cdk-log-to-s3.LogBucket.grantPut.parameter.identity"></a>

- *Type:* aws-cdk-lib.aws_iam.IGrantable

The principal.

---

###### `objectsKeyPattern`<sup>Optional</sup> <a name="objectsKeyPattern" id="@jaggr2/cdk-log-to-s3.LogBucket.grantPut.parameter.objectsKeyPattern"></a>

- *Type:* any

Restrict the permission to a certain key pattern (default '*').

Parameter type is `any` but `string` should be passed in.

---

##### `grantPutAcl` <a name="grantPutAcl" id="@jaggr2/cdk-log-to-s3.LogBucket.grantPutAcl"></a>

```typescript
public grantPutAcl(identity: IGrantable, objectsKeyPattern?: string): Grant
```

Grant the given IAM identity permissions to modify the ACLs of objects in the given Bucket.

If your application has the '@aws-cdk/aws-s3:grantWriteWithoutAcl' feature flag set,
calling `grantWrite` or `grantReadWrite` no longer grants permissions to modify the ACLs of the objects;
in this case, if you need to modify object ACLs, call this method explicitly.

###### `identity`<sup>Required</sup> <a name="identity" id="@jaggr2/cdk-log-to-s3.LogBucket.grantPutAcl.parameter.identity"></a>

- *Type:* aws-cdk-lib.aws_iam.IGrantable

---

###### `objectsKeyPattern`<sup>Optional</sup> <a name="objectsKeyPattern" id="@jaggr2/cdk-log-to-s3.LogBucket.grantPutAcl.parameter.objectsKeyPattern"></a>

- *Type:* string

---

##### `grantRead` <a name="grantRead" id="@jaggr2/cdk-log-to-s3.LogBucket.grantRead"></a>

```typescript
public grantRead(identity: IGrantable, objectsKeyPattern?: any): Grant
```

Grant read permissions for this bucket and it's contents to an IAM principal (Role/Group/User).

If encryption is used, permission to use the key to decrypt the contents
of the bucket will also be granted to the same principal.

###### `identity`<sup>Required</sup> <a name="identity" id="@jaggr2/cdk-log-to-s3.LogBucket.grantRead.parameter.identity"></a>

- *Type:* aws-cdk-lib.aws_iam.IGrantable

The principal.

---

###### `objectsKeyPattern`<sup>Optional</sup> <a name="objectsKeyPattern" id="@jaggr2/cdk-log-to-s3.LogBucket.grantRead.parameter.objectsKeyPattern"></a>

- *Type:* any

Restrict the permission to a certain key pattern (default '*').

Parameter type is `any` but `string` should be passed in.

---

##### `grantReadWrite` <a name="grantReadWrite" id="@jaggr2/cdk-log-to-s3.LogBucket.grantReadWrite"></a>

```typescript
public grantReadWrite(identity: IGrantable, objectsKeyPattern?: any): Grant
```

Grants read/write permissions for this bucket and it's contents to an IAM principal (Role/Group/User).

If an encryption key is used, permission to use the key for
encrypt/decrypt will also be granted.

Before CDK version 1.85.0, this method granted the `s3:PutObject*` permission that included `s3:PutObjectAcl`,
which could be used to grant read/write object access to IAM principals in other accounts.
If you want to get rid of that behavior, update your CDK version to 1.85.0 or later,
and make sure the `@aws-cdk/aws-s3:grantWriteWithoutAcl` feature flag is set to `true`
in the `context` key of your cdk.json file.
If you've already updated, but still need the principal to have permissions to modify the ACLs,
use the `grantPutAcl` method.

###### `identity`<sup>Required</sup> <a name="identity" id="@jaggr2/cdk-log-to-s3.LogBucket.grantReadWrite.parameter.identity"></a>

- *Type:* aws-cdk-lib.aws_iam.IGrantable

---

###### `objectsKeyPattern`<sup>Optional</sup> <a name="objectsKeyPattern" id="@jaggr2/cdk-log-to-s3.LogBucket.grantReadWrite.parameter.objectsKeyPattern"></a>

- *Type:* any

---

##### `grantWrite` <a name="grantWrite" id="@jaggr2/cdk-log-to-s3.LogBucket.grantWrite"></a>

```typescript
public grantWrite(identity: IGrantable, objectsKeyPattern?: any, allowedActionPatterns?: string[]): Grant
```

Grant write permissions to this bucket to an IAM principal.

If encryption is used, permission to use the key to encrypt the contents
of written files will also be granted to the same principal.

Before CDK version 1.85.0, this method granted the `s3:PutObject*` permission that included `s3:PutObjectAcl`,
which could be used to grant read/write object access to IAM principals in other accounts.
If you want to get rid of that behavior, update your CDK version to 1.85.0 or later,
and make sure the `@aws-cdk/aws-s3:grantWriteWithoutAcl` feature flag is set to `true`
in the `context` key of your cdk.json file.
If you've already updated, but still need the principal to have permissions to modify the ACLs,
use the `grantPutAcl` method.

###### `identity`<sup>Required</sup> <a name="identity" id="@jaggr2/cdk-log-to-s3.LogBucket.grantWrite.parameter.identity"></a>

- *Type:* aws-cdk-lib.aws_iam.IGrantable

---

###### `objectsKeyPattern`<sup>Optional</sup> <a name="objectsKeyPattern" id="@jaggr2/cdk-log-to-s3.LogBucket.grantWrite.parameter.objectsKeyPattern"></a>

- *Type:* any

---

###### `allowedActionPatterns`<sup>Optional</sup> <a name="allowedActionPatterns" id="@jaggr2/cdk-log-to-s3.LogBucket.grantWrite.parameter.allowedActionPatterns"></a>

- *Type:* string[]

---

##### `onCloudTrailEvent` <a name="onCloudTrailEvent" id="@jaggr2/cdk-log-to-s3.LogBucket.onCloudTrailEvent"></a>

```typescript
public onCloudTrailEvent(id: string, options?: OnCloudTrailBucketEventOptions): Rule
```

Define a CloudWatch event that triggers when something happens to this repository.

Requires that there exists at least one CloudTrail Trail in your account
that captures the event. This method will not create the Trail.

###### `id`<sup>Required</sup> <a name="id" id="@jaggr2/cdk-log-to-s3.LogBucket.onCloudTrailEvent.parameter.id"></a>

- *Type:* string

The id of the rule.

---

###### `options`<sup>Optional</sup> <a name="options" id="@jaggr2/cdk-log-to-s3.LogBucket.onCloudTrailEvent.parameter.options"></a>

- *Type:* aws-cdk-lib.aws_s3.OnCloudTrailBucketEventOptions

Options for adding the rule.

---

##### `onCloudTrailPutObject` <a name="onCloudTrailPutObject" id="@jaggr2/cdk-log-to-s3.LogBucket.onCloudTrailPutObject"></a>

```typescript
public onCloudTrailPutObject(id: string, options?: OnCloudTrailBucketEventOptions): Rule
```

Defines an AWS CloudWatch event that triggers when an object is uploaded to the specified paths (keys) in this bucket using the PutObject API call.

Note that some tools like `aws s3 cp` will automatically use either
PutObject or the multipart upload API depending on the file size,
so using `onCloudTrailWriteObject` may be preferable.

Requires that there exists at least one CloudTrail Trail in your account
that captures the event. This method will not create the Trail.

###### `id`<sup>Required</sup> <a name="id" id="@jaggr2/cdk-log-to-s3.LogBucket.onCloudTrailPutObject.parameter.id"></a>

- *Type:* string

The id of the rule.

---

###### `options`<sup>Optional</sup> <a name="options" id="@jaggr2/cdk-log-to-s3.LogBucket.onCloudTrailPutObject.parameter.options"></a>

- *Type:* aws-cdk-lib.aws_s3.OnCloudTrailBucketEventOptions

Options for adding the rule.

---

##### `onCloudTrailWriteObject` <a name="onCloudTrailWriteObject" id="@jaggr2/cdk-log-to-s3.LogBucket.onCloudTrailWriteObject"></a>

```typescript
public onCloudTrailWriteObject(id: string, options?: OnCloudTrailBucketEventOptions): Rule
```

Defines an AWS CloudWatch event that triggers when an object at the specified paths (keys) in this bucket are written to.

This includes
the events PutObject, CopyObject, and CompleteMultipartUpload.

Note that some tools like `aws s3 cp` will automatically use either
PutObject or the multipart upload API depending on the file size,
so using this method may be preferable to `onCloudTrailPutObject`.

Requires that there exists at least one CloudTrail Trail in your account
that captures the event. This method will not create the Trail.

###### `id`<sup>Required</sup> <a name="id" id="@jaggr2/cdk-log-to-s3.LogBucket.onCloudTrailWriteObject.parameter.id"></a>

- *Type:* string

The id of the rule.

---

###### `options`<sup>Optional</sup> <a name="options" id="@jaggr2/cdk-log-to-s3.LogBucket.onCloudTrailWriteObject.parameter.options"></a>

- *Type:* aws-cdk-lib.aws_s3.OnCloudTrailBucketEventOptions

Options for adding the rule.

---

##### `s3UrlForObject` <a name="s3UrlForObject" id="@jaggr2/cdk-log-to-s3.LogBucket.s3UrlForObject"></a>

```typescript
public s3UrlForObject(key?: string): string
```

The S3 URL of an S3 object. For example:.

`s3://onlybucket`
- `s3://bucket/key`

###### `key`<sup>Optional</sup> <a name="key" id="@jaggr2/cdk-log-to-s3.LogBucket.s3UrlForObject.parameter.key"></a>

- *Type:* string

The S3 key of the object.

If not specified, the S3 URL of the
bucket is returned.

---

##### `transferAccelerationUrlForObject` <a name="transferAccelerationUrlForObject" id="@jaggr2/cdk-log-to-s3.LogBucket.transferAccelerationUrlForObject"></a>

```typescript
public transferAccelerationUrlForObject(key?: string, options?: TransferAccelerationUrlOptions): string
```

The https Transfer Acceleration URL of an S3 object.

Specify `dualStack: true` at the options
for dual-stack endpoint (connect to the bucket over IPv6). For example:

- `https://bucket.s3-accelerate.amazonaws.com`
- `https://bucket.s3-accelerate.amazonaws.com/key`

###### `key`<sup>Optional</sup> <a name="key" id="@jaggr2/cdk-log-to-s3.LogBucket.transferAccelerationUrlForObject.parameter.key"></a>

- *Type:* string

The S3 key of the object.

If not specified, the URL of the
bucket is returned.

---

###### `options`<sup>Optional</sup> <a name="options" id="@jaggr2/cdk-log-to-s3.LogBucket.transferAccelerationUrlForObject.parameter.options"></a>

- *Type:* aws-cdk-lib.aws_s3.TransferAccelerationUrlOptions

Options for generating URL.

---

##### `urlForObject` <a name="urlForObject" id="@jaggr2/cdk-log-to-s3.LogBucket.urlForObject"></a>

```typescript
public urlForObject(key?: string): string
```

The https URL of an S3 object. Specify `regional: false` at the options for non-regional URLs. For example:.

`https://s3.us-west-1.amazonaws.com/onlybucket`
- `https://s3.us-west-1.amazonaws.com/bucket/key`
- `https://s3.cn-north-1.amazonaws.com.cn/china-bucket/mykey`

###### `key`<sup>Optional</sup> <a name="key" id="@jaggr2/cdk-log-to-s3.LogBucket.urlForObject.parameter.key"></a>

- *Type:* string

The S3 key of the object.

If not specified, the URL of the
bucket is returned.

---

##### `virtualHostedUrlForObject` <a name="virtualHostedUrlForObject" id="@jaggr2/cdk-log-to-s3.LogBucket.virtualHostedUrlForObject"></a>

```typescript
public virtualHostedUrlForObject(key?: string, options?: VirtualHostedStyleUrlOptions): string
```

The virtual hosted-style URL of an S3 object. Specify `regional: false` at the options for non-regional URL. For example:.

`https://only-bucket.s3.us-west-1.amazonaws.com`
- `https://bucket.s3.us-west-1.amazonaws.com/key`
- `https://bucket.s3.amazonaws.com/key`
- `https://china-bucket.s3.cn-north-1.amazonaws.com.cn/mykey`

###### `key`<sup>Optional</sup> <a name="key" id="@jaggr2/cdk-log-to-s3.LogBucket.virtualHostedUrlForObject.parameter.key"></a>

- *Type:* string

The S3 key of the object.

If not specified, the URL of the
bucket is returned.

---

###### `options`<sup>Optional</sup> <a name="options" id="@jaggr2/cdk-log-to-s3.LogBucket.virtualHostedUrlForObject.parameter.options"></a>

- *Type:* aws-cdk-lib.aws_s3.VirtualHostedStyleUrlOptions

Options for generating URL.

---

##### `addCorsRule` <a name="addCorsRule" id="@jaggr2/cdk-log-to-s3.LogBucket.addCorsRule"></a>

```typescript
public addCorsRule(rule: CorsRule): void
```

Adds a cross-origin access configuration for objects in an Amazon S3 bucket.

###### `rule`<sup>Required</sup> <a name="rule" id="@jaggr2/cdk-log-to-s3.LogBucket.addCorsRule.parameter.rule"></a>

- *Type:* aws-cdk-lib.aws_s3.CorsRule

The CORS configuration rule to add.

---

##### `addInventory` <a name="addInventory" id="@jaggr2/cdk-log-to-s3.LogBucket.addInventory"></a>

```typescript
public addInventory(inventory: Inventory): void
```

Add an inventory configuration.

###### `inventory`<sup>Required</sup> <a name="inventory" id="@jaggr2/cdk-log-to-s3.LogBucket.addInventory.parameter.inventory"></a>

- *Type:* aws-cdk-lib.aws_s3.Inventory

configuration to add.

---

##### `addLifecycleRule` <a name="addLifecycleRule" id="@jaggr2/cdk-log-to-s3.LogBucket.addLifecycleRule"></a>

```typescript
public addLifecycleRule(rule: LifecycleRule): void
```

Add a lifecycle rule to the bucket.

###### `rule`<sup>Required</sup> <a name="rule" id="@jaggr2/cdk-log-to-s3.LogBucket.addLifecycleRule.parameter.rule"></a>

- *Type:* aws-cdk-lib.aws_s3.LifecycleRule

The rule to add.

---

##### `addMetric` <a name="addMetric" id="@jaggr2/cdk-log-to-s3.LogBucket.addMetric"></a>

```typescript
public addMetric(metric: BucketMetrics): void
```

Adds a metrics configuration for the CloudWatch request metrics from the bucket.

###### `metric`<sup>Required</sup> <a name="metric" id="@jaggr2/cdk-log-to-s3.LogBucket.addMetric.parameter.metric"></a>

- *Type:* aws-cdk-lib.aws_s3.BucketMetrics

The metric configuration to add.

---

#### Static Functions <a name="Static Functions" id="Static Functions"></a>

| **Name** | **Description** |
| --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.isConstruct">isConstruct</a></code> | Checks if `x` is a construct. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.isOwnedResource">isOwnedResource</a></code> | Returns true if the construct was created by CDK, and false otherwise. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.isResource">isResource</a></code> | Check whether the given construct is a Resource. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.fromBucketArn">fromBucketArn</a></code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.fromBucketAttributes">fromBucketAttributes</a></code> | Creates a Bucket construct that represents an external bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.fromBucketName">fromBucketName</a></code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.fromCfnBucket">fromCfnBucket</a></code> | Create a mutable `IBucket` based on a low-level `CfnBucket`. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.validateBucketName">validateBucketName</a></code> | Thrown an exception if the given bucket name is not valid. |

---

##### `isConstruct` <a name="isConstruct" id="@jaggr2/cdk-log-to-s3.LogBucket.isConstruct"></a>

```typescript
import { LogBucket } from '@jaggr2/cdk-log-to-s3'

LogBucket.isConstruct(x: any)
```

Checks if `x` is a construct.

Use this method instead of `instanceof` to properly detect `Construct`
instances, even when the construct library is symlinked.

Explanation: in JavaScript, multiple copies of the `constructs` library on
disk are seen as independent, completely different libraries. As a
consequence, the class `Construct` in each copy of the `constructs` library
is seen as a different class, and an instance of one class will not test as
`instanceof` the other class. `npm install` will not create installations
like this, but users may manually symlink construct libraries together or
use a monorepo tool: in those cases, multiple copies of the `constructs`
library can be accidentally installed, and `instanceof` will behave
unpredictably. It is safest to avoid using `instanceof`, and using
this type-testing method instead.

###### `x`<sup>Required</sup> <a name="x" id="@jaggr2/cdk-log-to-s3.LogBucket.isConstruct.parameter.x"></a>

- *Type:* any

Any object.

---

##### `isOwnedResource` <a name="isOwnedResource" id="@jaggr2/cdk-log-to-s3.LogBucket.isOwnedResource"></a>

```typescript
import { LogBucket } from '@jaggr2/cdk-log-to-s3'

LogBucket.isOwnedResource(construct: IConstruct)
```

Returns true if the construct was created by CDK, and false otherwise.

###### `construct`<sup>Required</sup> <a name="construct" id="@jaggr2/cdk-log-to-s3.LogBucket.isOwnedResource.parameter.construct"></a>

- *Type:* constructs.IConstruct

---

##### `isResource` <a name="isResource" id="@jaggr2/cdk-log-to-s3.LogBucket.isResource"></a>

```typescript
import { LogBucket } from '@jaggr2/cdk-log-to-s3'

LogBucket.isResource(construct: IConstruct)
```

Check whether the given construct is a Resource.

###### `construct`<sup>Required</sup> <a name="construct" id="@jaggr2/cdk-log-to-s3.LogBucket.isResource.parameter.construct"></a>

- *Type:* constructs.IConstruct

---

##### `fromBucketArn` <a name="fromBucketArn" id="@jaggr2/cdk-log-to-s3.LogBucket.fromBucketArn"></a>

```typescript
import { LogBucket } from '@jaggr2/cdk-log-to-s3'

LogBucket.fromBucketArn(scope: Construct, id: string, bucketArn: string)
```

###### `scope`<sup>Required</sup> <a name="scope" id="@jaggr2/cdk-log-to-s3.LogBucket.fromBucketArn.parameter.scope"></a>

- *Type:* constructs.Construct

---

###### `id`<sup>Required</sup> <a name="id" id="@jaggr2/cdk-log-to-s3.LogBucket.fromBucketArn.parameter.id"></a>

- *Type:* string

---

###### `bucketArn`<sup>Required</sup> <a name="bucketArn" id="@jaggr2/cdk-log-to-s3.LogBucket.fromBucketArn.parameter.bucketArn"></a>

- *Type:* string

---

##### `fromBucketAttributes` <a name="fromBucketAttributes" id="@jaggr2/cdk-log-to-s3.LogBucket.fromBucketAttributes"></a>

```typescript
import { LogBucket } from '@jaggr2/cdk-log-to-s3'

LogBucket.fromBucketAttributes(scope: Construct, id: string, attrs: BucketAttributes)
```

Creates a Bucket construct that represents an external bucket.

###### `scope`<sup>Required</sup> <a name="scope" id="@jaggr2/cdk-log-to-s3.LogBucket.fromBucketAttributes.parameter.scope"></a>

- *Type:* constructs.Construct

The parent creating construct (usually `this`).

---

###### `id`<sup>Required</sup> <a name="id" id="@jaggr2/cdk-log-to-s3.LogBucket.fromBucketAttributes.parameter.id"></a>

- *Type:* string

The construct's name.

---

###### `attrs`<sup>Required</sup> <a name="attrs" id="@jaggr2/cdk-log-to-s3.LogBucket.fromBucketAttributes.parameter.attrs"></a>

- *Type:* aws-cdk-lib.aws_s3.BucketAttributes

A `BucketAttributes` object.

Can be obtained from a call to
`bucket.export()` or manually created.

---

##### `fromBucketName` <a name="fromBucketName" id="@jaggr2/cdk-log-to-s3.LogBucket.fromBucketName"></a>

```typescript
import { LogBucket } from '@jaggr2/cdk-log-to-s3'

LogBucket.fromBucketName(scope: Construct, id: string, bucketName: string)
```

###### `scope`<sup>Required</sup> <a name="scope" id="@jaggr2/cdk-log-to-s3.LogBucket.fromBucketName.parameter.scope"></a>

- *Type:* constructs.Construct

---

###### `id`<sup>Required</sup> <a name="id" id="@jaggr2/cdk-log-to-s3.LogBucket.fromBucketName.parameter.id"></a>

- *Type:* string

---

###### `bucketName`<sup>Required</sup> <a name="bucketName" id="@jaggr2/cdk-log-to-s3.LogBucket.fromBucketName.parameter.bucketName"></a>

- *Type:* string

---

##### `fromCfnBucket` <a name="fromCfnBucket" id="@jaggr2/cdk-log-to-s3.LogBucket.fromCfnBucket"></a>

```typescript
import { LogBucket } from '@jaggr2/cdk-log-to-s3'

LogBucket.fromCfnBucket(cfnBucket: CfnBucket)
```

Create a mutable `IBucket` based on a low-level `CfnBucket`.

###### `cfnBucket`<sup>Required</sup> <a name="cfnBucket" id="@jaggr2/cdk-log-to-s3.LogBucket.fromCfnBucket.parameter.cfnBucket"></a>

- *Type:* aws-cdk-lib.aws_s3.CfnBucket

---

##### `validateBucketName` <a name="validateBucketName" id="@jaggr2/cdk-log-to-s3.LogBucket.validateBucketName"></a>

```typescript
import { LogBucket } from '@jaggr2/cdk-log-to-s3'

LogBucket.validateBucketName(physicalName: string, allowLegacyBucketNaming?: boolean)
```

Thrown an exception if the given bucket name is not valid.

###### `physicalName`<sup>Required</sup> <a name="physicalName" id="@jaggr2/cdk-log-to-s3.LogBucket.validateBucketName.parameter.physicalName"></a>

- *Type:* string

name of the bucket.

---

###### `allowLegacyBucketNaming`<sup>Optional</sup> <a name="allowLegacyBucketNaming" id="@jaggr2/cdk-log-to-s3.LogBucket.validateBucketName.parameter.allowLegacyBucketNaming"></a>

- *Type:* boolean

allow legacy bucket naming style, default is false.

---

#### Properties <a name="Properties" id="Properties"></a>

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.property.node">node</a></code> | <code>constructs.Node</code> | The tree node. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.property.env">env</a></code> | <code>aws-cdk-lib.ResourceEnvironment</code> | The environment this resource belongs to. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.property.stack">stack</a></code> | <code>aws-cdk-lib.Stack</code> | The stack in which this resource is defined. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.property.bucketArn">bucketArn</a></code> | <code>string</code> | The ARN of the bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.property.bucketDomainName">bucketDomainName</a></code> | <code>string</code> | The IPv4 DNS name of the specified bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.property.bucketDualStackDomainName">bucketDualStackDomainName</a></code> | <code>string</code> | The IPv6 DNS name of the specified bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.property.bucketName">bucketName</a></code> | <code>string</code> | The name of the bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.property.bucketRegionalDomainName">bucketRegionalDomainName</a></code> | <code>string</code> | The regional domain name of the specified bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.property.bucketWebsiteDomainName">bucketWebsiteDomainName</a></code> | <code>string</code> | The Domain name of the static website. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.property.bucketWebsiteUrl">bucketWebsiteUrl</a></code> | <code>string</code> | The URL of the static website. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.property.encryptionKey">encryptionKey</a></code> | <code>aws-cdk-lib.aws_kms.IKey</code> | Optional KMS encryption key associated with this bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.property.isWebsite">isWebsite</a></code> | <code>boolean</code> | If this bucket has been configured for static website hosting. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.property.policy">policy</a></code> | <code>aws-cdk-lib.aws_s3.BucketPolicy</code> | The resource policy associated with this bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucket.property.keyPrefix">keyPrefix</a></code> | <code>string</code> | Normalised key prefix, e.g. 'logs/'. |

---

##### `node`<sup>Required</sup> <a name="node" id="@jaggr2/cdk-log-to-s3.LogBucket.property.node"></a>

```typescript
public readonly node: Node;
```

- *Type:* constructs.Node

The tree node.

---

##### `env`<sup>Required</sup> <a name="env" id="@jaggr2/cdk-log-to-s3.LogBucket.property.env"></a>

```typescript
public readonly env: ResourceEnvironment;
```

- *Type:* aws-cdk-lib.ResourceEnvironment

The environment this resource belongs to.

For resources that are created and managed by the CDK
(generally, those created by creating new class instances like Role, Bucket, etc.),
this is always the same as the environment of the stack they belong to;
however, for imported resources
(those obtained from static methods like fromRoleArn, fromBucketName, etc.),
that might be different than the stack they were imported into.

---

##### `stack`<sup>Required</sup> <a name="stack" id="@jaggr2/cdk-log-to-s3.LogBucket.property.stack"></a>

```typescript
public readonly stack: Stack;
```

- *Type:* aws-cdk-lib.Stack

The stack in which this resource is defined.

---

##### `bucketArn`<sup>Required</sup> <a name="bucketArn" id="@jaggr2/cdk-log-to-s3.LogBucket.property.bucketArn"></a>

```typescript
public readonly bucketArn: string;
```

- *Type:* string

The ARN of the bucket.

---

##### `bucketDomainName`<sup>Required</sup> <a name="bucketDomainName" id="@jaggr2/cdk-log-to-s3.LogBucket.property.bucketDomainName"></a>

```typescript
public readonly bucketDomainName: string;
```

- *Type:* string

The IPv4 DNS name of the specified bucket.

---

##### `bucketDualStackDomainName`<sup>Required</sup> <a name="bucketDualStackDomainName" id="@jaggr2/cdk-log-to-s3.LogBucket.property.bucketDualStackDomainName"></a>

```typescript
public readonly bucketDualStackDomainName: string;
```

- *Type:* string

The IPv6 DNS name of the specified bucket.

---

##### `bucketName`<sup>Required</sup> <a name="bucketName" id="@jaggr2/cdk-log-to-s3.LogBucket.property.bucketName"></a>

```typescript
public readonly bucketName: string;
```

- *Type:* string

The name of the bucket.

---

##### `bucketRegionalDomainName`<sup>Required</sup> <a name="bucketRegionalDomainName" id="@jaggr2/cdk-log-to-s3.LogBucket.property.bucketRegionalDomainName"></a>

```typescript
public readonly bucketRegionalDomainName: string;
```

- *Type:* string

The regional domain name of the specified bucket.

---

##### `bucketWebsiteDomainName`<sup>Required</sup> <a name="bucketWebsiteDomainName" id="@jaggr2/cdk-log-to-s3.LogBucket.property.bucketWebsiteDomainName"></a>

```typescript
public readonly bucketWebsiteDomainName: string;
```

- *Type:* string

The Domain name of the static website.

---

##### `bucketWebsiteUrl`<sup>Required</sup> <a name="bucketWebsiteUrl" id="@jaggr2/cdk-log-to-s3.LogBucket.property.bucketWebsiteUrl"></a>

```typescript
public readonly bucketWebsiteUrl: string;
```

- *Type:* string

The URL of the static website.

---

##### `encryptionKey`<sup>Optional</sup> <a name="encryptionKey" id="@jaggr2/cdk-log-to-s3.LogBucket.property.encryptionKey"></a>

```typescript
public readonly encryptionKey: IKey;
```

- *Type:* aws-cdk-lib.aws_kms.IKey

Optional KMS encryption key associated with this bucket.

---

##### `isWebsite`<sup>Optional</sup> <a name="isWebsite" id="@jaggr2/cdk-log-to-s3.LogBucket.property.isWebsite"></a>

```typescript
public readonly isWebsite: boolean;
```

- *Type:* boolean

If this bucket has been configured for static website hosting.

---

##### `policy`<sup>Optional</sup> <a name="policy" id="@jaggr2/cdk-log-to-s3.LogBucket.property.policy"></a>

```typescript
public readonly policy: BucketPolicy;
```

- *Type:* aws-cdk-lib.aws_s3.BucketPolicy

The resource policy associated with this bucket.

If `autoCreatePolicy` is true, a `BucketPolicy` will be created upon the
first call to addToResourcePolicy(s).

---

##### `keyPrefix`<sup>Required</sup> <a name="keyPrefix" id="@jaggr2/cdk-log-to-s3.LogBucket.property.keyPrefix"></a>

```typescript
public readonly keyPrefix: string;
```

- *Type:* string

Normalised key prefix, e.g. 'logs/'.

---


### LogCompaction <a name="LogCompaction" id="@jaggr2/cdk-log-to-s3.LogCompaction"></a>

A daily job that merges the many small Parquet files the extension produces into fewer, larger ones.

The extension flushes on a timer, on a size threshold and at the end of every
invocation, so a busy function can leave thousands of tiny objects in a day
partition. Athena pays a per-file cost opening footers, so that is slow to
scan regardless of how little data it holds. Compaction is the answer to that
- not finer partitioning, which only moves the cost into the query planner.

It works purely at the S3 level and never touches the Glue catalog. Under
partition projection there is nothing to register: Athena computes partitions
from the `dt` range at query time, and rewriting files inside a partition
does not change the partition set. This is why the construct needs no Glue
permissions and why `MSCK REPAIR` has no role here.

Only closed days are compacted; today is left alone while the extension is
still writing into it.

*Example*

```typescript
const compaction = LogCompaction.fromExtension(this, 'Compaction', extension);
```


#### Initializers <a name="Initializers" id="@jaggr2/cdk-log-to-s3.LogCompaction.Initializer"></a>

```typescript
import { LogCompaction } from '@jaggr2/cdk-log-to-s3'

new LogCompaction(scope: Construct, id: string, props: LogCompactionProps)
```

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompaction.Initializer.parameter.scope">scope</a></code> | <code>constructs.Construct</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompaction.Initializer.parameter.id">id</a></code> | <code>string</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompaction.Initializer.parameter.props">props</a></code> | <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps">LogCompactionProps</a></code> | *No description.* |

---

##### `scope`<sup>Required</sup> <a name="scope" id="@jaggr2/cdk-log-to-s3.LogCompaction.Initializer.parameter.scope"></a>

- *Type:* constructs.Construct

---

##### `id`<sup>Required</sup> <a name="id" id="@jaggr2/cdk-log-to-s3.LogCompaction.Initializer.parameter.id"></a>

- *Type:* string

---

##### `props`<sup>Required</sup> <a name="props" id="@jaggr2/cdk-log-to-s3.LogCompaction.Initializer.parameter.props"></a>

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps">LogCompactionProps</a>

---

#### Methods <a name="Methods" id="Methods"></a>

| **Name** | **Description** |
| --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompaction.toString">toString</a></code> | Returns a string representation of this construct. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompaction.with">with</a></code> | Applies one or more mixins to this construct. |

---

##### `toString` <a name="toString" id="@jaggr2/cdk-log-to-s3.LogCompaction.toString"></a>

```typescript
public toString(): string
```

Returns a string representation of this construct.

##### `with` <a name="with" id="@jaggr2/cdk-log-to-s3.LogCompaction.with"></a>

```typescript
public with(mixins: ...IMixin[]): IConstruct
```

Applies one or more mixins to this construct.

Mixins are applied in order. The list of constructs is captured at the
start of the call, so constructs added by a mixin will not be visited.
Use multiple `with()` calls if subsequent mixins should apply to added
constructs.

###### `mixins`<sup>Required</sup> <a name="mixins" id="@jaggr2/cdk-log-to-s3.LogCompaction.with.parameter.mixins"></a>

- *Type:* ...constructs.IMixin[]

The mixins to apply.

---

#### Static Functions <a name="Static Functions" id="Static Functions"></a>

| **Name** | **Description** |
| --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompaction.isConstruct">isConstruct</a></code> | Checks if `x` is a construct. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompaction.fromExtension">fromExtension</a></code> | Build the job for an extension, taking the bucket, key prefix, architecture and codec from it so they cannot drift apart. |

---

##### `isConstruct` <a name="isConstruct" id="@jaggr2/cdk-log-to-s3.LogCompaction.isConstruct"></a>

```typescript
import { LogCompaction } from '@jaggr2/cdk-log-to-s3'

LogCompaction.isConstruct(x: any)
```

Checks if `x` is a construct.

Use this method instead of `instanceof` to properly detect `Construct`
instances, even when the construct library is symlinked.

Explanation: in JavaScript, multiple copies of the `constructs` library on
disk are seen as independent, completely different libraries. As a
consequence, the class `Construct` in each copy of the `constructs` library
is seen as a different class, and an instance of one class will not test as
`instanceof` the other class. `npm install` will not create installations
like this, but users may manually symlink construct libraries together or
use a monorepo tool: in those cases, multiple copies of the `constructs`
library can be accidentally installed, and `instanceof` will behave
unpredictably. It is safest to avoid using `instanceof`, and using
this type-testing method instead.

###### `x`<sup>Required</sup> <a name="x" id="@jaggr2/cdk-log-to-s3.LogCompaction.isConstruct.parameter.x"></a>

- *Type:* any

Any object.

---

##### `fromExtension` <a name="fromExtension" id="@jaggr2/cdk-log-to-s3.LogCompaction.fromExtension"></a>

```typescript
import { LogCompaction } from '@jaggr2/cdk-log-to-s3'

LogCompaction.fromExtension(scope: Construct, id: string, extension: ILogToS3Extension, options?: LogCompactionFromExtensionOptions)
```

Build the job for an extension, taking the bucket, key prefix, architecture and codec from it so they cannot drift apart.

###### `scope`<sup>Required</sup> <a name="scope" id="@jaggr2/cdk-log-to-s3.LogCompaction.fromExtension.parameter.scope"></a>

- *Type:* constructs.Construct

---

###### `id`<sup>Required</sup> <a name="id" id="@jaggr2/cdk-log-to-s3.LogCompaction.fromExtension.parameter.id"></a>

- *Type:* string

---

###### `extension`<sup>Required</sup> <a name="extension" id="@jaggr2/cdk-log-to-s3.LogCompaction.fromExtension.parameter.extension"></a>

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.ILogToS3Extension">ILogToS3Extension</a>

---

###### `options`<sup>Optional</sup> <a name="options" id="@jaggr2/cdk-log-to-s3.LogCompaction.fromExtension.parameter.options"></a>

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions">LogCompactionFromExtensionOptions</a>

---

#### Properties <a name="Properties" id="Properties"></a>

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompaction.property.node">node</a></code> | <code>constructs.Node</code> | The tree node. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompaction.property.functionArn">functionArn</a></code> | <code>string</code> | ARN of the compaction function, for wiring alarms or manual invocation. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompaction.property.handler">handler</a></code> | <code>aws-cdk-lib.aws_lambda.Function</code> | The compaction function. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompaction.property.keyPrefix">keyPrefix</a></code> | <code>string</code> | Normalised key prefix the job operates under. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompaction.property.rule">rule</a></code> | <code>aws-cdk-lib.aws_events.Rule</code> | The schedule, or undefined when `enabled` was false. |

---

##### `node`<sup>Required</sup> <a name="node" id="@jaggr2/cdk-log-to-s3.LogCompaction.property.node"></a>

```typescript
public readonly node: Node;
```

- *Type:* constructs.Node

The tree node.

---

##### `functionArn`<sup>Required</sup> <a name="functionArn" id="@jaggr2/cdk-log-to-s3.LogCompaction.property.functionArn"></a>

```typescript
public readonly functionArn: string;
```

- *Type:* string

ARN of the compaction function, for wiring alarms or manual invocation.

---

##### `handler`<sup>Required</sup> <a name="handler" id="@jaggr2/cdk-log-to-s3.LogCompaction.property.handler"></a>

```typescript
public readonly handler: Function;
```

- *Type:* aws-cdk-lib.aws_lambda.Function

The compaction function.

---

##### `keyPrefix`<sup>Required</sup> <a name="keyPrefix" id="@jaggr2/cdk-log-to-s3.LogCompaction.property.keyPrefix"></a>

```typescript
public readonly keyPrefix: string;
```

- *Type:* string

Normalised key prefix the job operates under.

---

##### `rule`<sup>Optional</sup> <a name="rule" id="@jaggr2/cdk-log-to-s3.LogCompaction.property.rule"></a>

```typescript
public readonly rule: Rule;
```

- *Type:* aws-cdk-lib.aws_events.Rule

The schedule, or undefined when `enabled` was false.

---


### LogToS3Extension <a name="LogToS3Extension" id="@jaggr2/cdk-log-to-s3.LogToS3Extension"></a>

- *Implements:* <a href="#@jaggr2/cdk-log-to-s3.ILogToS3Extension">ILogToS3Extension</a>

A Lambda layer containing an external extension that subscribes to the Telemetry API and writes structured logs to S3 as Parquet.

Instantiate one per stack that owns functions, all pointing at the same
bucket. The layer asset hash is identical across stacks, so the zip is
uploaded once and the extra AWS::Lambda::LayerVersion resources are free.
That is deliberately cheaper than exporting the ARN across stacks, which
creates a CloudFormation export that cannot be changed while it is in use.

*Example*

```typescript
const bucket = new LogBucket(this, 'Logs');
const ext = new LogToS3Extension(this, 'LogExt', { logsBucket: bucket });
ext.attachTo(myFunction);
```


#### Initializers <a name="Initializers" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.Initializer"></a>

```typescript
import { LogToS3Extension } from '@jaggr2/cdk-log-to-s3'

new LogToS3Extension(scope: Construct, id: string, props: LogToS3ExtensionProps)
```

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension.Initializer.parameter.scope">scope</a></code> | <code>constructs.Construct</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension.Initializer.parameter.id">id</a></code> | <code>string</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension.Initializer.parameter.props">props</a></code> | <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps">LogToS3ExtensionProps</a></code> | *No description.* |

---

##### `scope`<sup>Required</sup> <a name="scope" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.Initializer.parameter.scope"></a>

- *Type:* constructs.Construct

---

##### `id`<sup>Required</sup> <a name="id" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.Initializer.parameter.id"></a>

- *Type:* string

---

##### `props`<sup>Required</sup> <a name="props" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.Initializer.parameter.props"></a>

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps">LogToS3ExtensionProps</a>

---

#### Methods <a name="Methods" id="Methods"></a>

| **Name** | **Description** |
| --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension.toString">toString</a></code> | Returns a string representation of this construct. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension.with">with</a></code> | Applies one or more mixins to this construct. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension.attachTo">attachTo</a></code> | Add the layer, the environment and the bucket grant to a function. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension.grantWriteLogs">grantWriteLogs</a></code> | Grant write access to the log prefix without attaching the layer. |

---

##### `toString` <a name="toString" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.toString"></a>

```typescript
public toString(): string
```

Returns a string representation of this construct.

##### `with` <a name="with" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.with"></a>

```typescript
public with(mixins: ...IMixin[]): IConstruct
```

Applies one or more mixins to this construct.

Mixins are applied in order. The list of constructs is captured at the
start of the call, so constructs added by a mixin will not be visited.
Use multiple `with()` calls if subsequent mixins should apply to added
constructs.

###### `mixins`<sup>Required</sup> <a name="mixins" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.with.parameter.mixins"></a>

- *Type:* ...constructs.IMixin[]

The mixins to apply.

---

##### `attachTo` <a name="attachTo" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.attachTo"></a>

```typescript
public attachTo(fn: Function, options?: LogToS3AttachOptions): void
```

Add the layer, the environment and the bucket grant to a function.

Takes a concrete lambda.Function rather than an IFunction because
addLayers() and addEnvironment() only exist on the concrete class.
Functions imported with Function.fromFunctionArn cannot be modified by
CloudFormation anyway.

###### `fn`<sup>Required</sup> <a name="fn" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.attachTo.parameter.fn"></a>

- *Type:* aws-cdk-lib.aws_lambda.Function

---

###### `options`<sup>Optional</sup> <a name="options" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.attachTo.parameter.options"></a>

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogToS3AttachOptions">LogToS3AttachOptions</a>

---

##### `grantWriteLogs` <a name="grantWriteLogs" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.grantWriteLogs"></a>

```typescript
public grantWriteLogs(grantee: IGrantable): Grant
```

Grant write access to the log prefix without attaching the layer.

###### `grantee`<sup>Required</sup> <a name="grantee" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.grantWriteLogs.parameter.grantee"></a>

- *Type:* aws-cdk-lib.aws_iam.IGrantable

---

#### Static Functions <a name="Static Functions" id="Static Functions"></a>

| **Name** | **Description** |
| --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension.isConstruct">isConstruct</a></code> | Checks if `x` is a construct. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension.fromAttributes">fromAttributes</a></code> | Reference a layer created by another app or stack. |

---

##### `isConstruct` <a name="isConstruct" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.isConstruct"></a>

```typescript
import { LogToS3Extension } from '@jaggr2/cdk-log-to-s3'

LogToS3Extension.isConstruct(x: any)
```

Checks if `x` is a construct.

Use this method instead of `instanceof` to properly detect `Construct`
instances, even when the construct library is symlinked.

Explanation: in JavaScript, multiple copies of the `constructs` library on
disk are seen as independent, completely different libraries. As a
consequence, the class `Construct` in each copy of the `constructs` library
is seen as a different class, and an instance of one class will not test as
`instanceof` the other class. `npm install` will not create installations
like this, but users may manually symlink construct libraries together or
use a monorepo tool: in those cases, multiple copies of the `constructs`
library can be accidentally installed, and `instanceof` will behave
unpredictably. It is safest to avoid using `instanceof`, and using
this type-testing method instead.

###### `x`<sup>Required</sup> <a name="x" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.isConstruct.parameter.x"></a>

- *Type:* any

Any object.

---

##### `fromAttributes` <a name="fromAttributes" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.fromAttributes"></a>

```typescript
import { LogToS3Extension } from '@jaggr2/cdk-log-to-s3'

LogToS3Extension.fromAttributes(scope: Construct, id: string, attrs: LogToS3ExtensionAttributes)
```

Reference a layer created by another app or stack.

###### `scope`<sup>Required</sup> <a name="scope" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.fromAttributes.parameter.scope"></a>

- *Type:* constructs.Construct

---

###### `id`<sup>Required</sup> <a name="id" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.fromAttributes.parameter.id"></a>

- *Type:* string

---

###### `attrs`<sup>Required</sup> <a name="attrs" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.fromAttributes.parameter.attrs"></a>

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes">LogToS3ExtensionAttributes</a>

---

#### Properties <a name="Properties" id="Properties"></a>

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension.property.node">node</a></code> | <code>constructs.Node</code> | The tree node. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension.property.architecture">architecture</a></code> | <code>aws-cdk-lib.aws_lambda.Architecture</code> | Architecture this layer was built for. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension.property.environment">environment</a></code> | <code>{[ key: string ]: string}</code> | Environment variables attachTo() injects. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension.property.keyPrefix">keyPrefix</a></code> | <code>string</code> | Normalised key prefix, e.g. 'logs/'. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension.property.layer">layer</a></code> | <code>aws-cdk-lib.aws_lambda.ILayerVersion</code> | The layer carrying the extension binary. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension.property.logsBucket">logsBucket</a></code> | <code>aws-cdk-lib.aws_s3.IBucket</code> | Bucket the extension writes to. |

---

##### `node`<sup>Required</sup> <a name="node" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.property.node"></a>

```typescript
public readonly node: Node;
```

- *Type:* constructs.Node

The tree node.

---

##### `architecture`<sup>Required</sup> <a name="architecture" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.property.architecture"></a>

```typescript
public readonly architecture: Architecture;
```

- *Type:* aws-cdk-lib.aws_lambda.Architecture

Architecture this layer was built for.

---

##### `environment`<sup>Required</sup> <a name="environment" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.property.environment"></a>

```typescript
public readonly environment: {[ key: string ]: string};
```

- *Type:* {[ key: string ]: string}

Environment variables attachTo() injects.

Exposed for manual wiring.

---

##### `keyPrefix`<sup>Required</sup> <a name="keyPrefix" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.property.keyPrefix"></a>

```typescript
public readonly keyPrefix: string;
```

- *Type:* string

Normalised key prefix, e.g. 'logs/'.

---

##### `layer`<sup>Required</sup> <a name="layer" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.property.layer"></a>

```typescript
public readonly layer: ILayerVersion;
```

- *Type:* aws-cdk-lib.aws_lambda.ILayerVersion

The layer carrying the extension binary.

---

##### `logsBucket`<sup>Required</sup> <a name="logsBucket" id="@jaggr2/cdk-log-to-s3.LogToS3Extension.property.logsBucket"></a>

```typescript
public readonly logsBucket: IBucket;
```

- *Type:* aws-cdk-lib.aws_s3.IBucket

Bucket the extension writes to.

---


## Structs <a name="Structs" id="Structs"></a>

### LogAnalyticsFromExtensionOptions <a name="LogAnalyticsFromExtensionOptions" id="@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions"></a>

Options for LogAnalytics.fromExtension.

#### Initializer <a name="Initializer" id="@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.Initializer"></a>

```typescript
import { LogAnalyticsFromExtensionOptions } from '@jaggr2/cdk-log-to-s3'

const logAnalyticsFromExtensionOptions: LogAnalyticsFromExtensionOptions = { ... }
```

#### Properties <a name="Properties" id="Properties"></a>

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.databaseName">databaseName</a></code> | <code>string</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.createDatabase">createDatabase</a></code> | <code>boolean</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.createWorkgroup">createWorkgroup</a></code> | <code>boolean</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.projectionWindow">projectionWindow</a></code> | <code>aws-cdk-lib.Duration</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.removalPolicy">removalPolicy</a></code> | <code>aws-cdk-lib.RemovalPolicy</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.resultsBucket">resultsBucket</a></code> | <code>aws-cdk-lib.aws_s3.IBucket</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.resultsExpiration">resultsExpiration</a></code> | <code>aws-cdk-lib.Duration</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.tableName">tableName</a></code> | <code>string</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.workgroupName">workgroupName</a></code> | <code>string</code> | *No description.* |

---

##### `databaseName`<sup>Required</sup> <a name="databaseName" id="@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.databaseName"></a>

```typescript
public readonly databaseName: string;
```

- *Type:* string

---

##### `createDatabase`<sup>Optional</sup> <a name="createDatabase" id="@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.createDatabase"></a>

```typescript
public readonly createDatabase: boolean;
```

- *Type:* boolean
- *Default:* true

---

##### `createWorkgroup`<sup>Optional</sup> <a name="createWorkgroup" id="@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.createWorkgroup"></a>

```typescript
public readonly createWorkgroup: boolean;
```

- *Type:* boolean
- *Default:* true

---

##### `projectionWindow`<sup>Optional</sup> <a name="projectionWindow" id="@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.projectionWindow"></a>

```typescript
public readonly projectionWindow: Duration;
```

- *Type:* aws-cdk-lib.Duration
- *Default:* Duration.days(730)

---

##### `removalPolicy`<sup>Optional</sup> <a name="removalPolicy" id="@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.removalPolicy"></a>

```typescript
public readonly removalPolicy: RemovalPolicy;
```

- *Type:* aws-cdk-lib.RemovalPolicy
- *Default:* RemovalPolicy.RETAIN

---

##### `resultsBucket`<sup>Optional</sup> <a name="resultsBucket" id="@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.resultsBucket"></a>

```typescript
public readonly resultsBucket: IBucket;
```

- *Type:* aws-cdk-lib.aws_s3.IBucket
- *Default:* a bucket is created with a 30 day expiry

---

##### `resultsExpiration`<sup>Optional</sup> <a name="resultsExpiration" id="@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.resultsExpiration"></a>

```typescript
public readonly resultsExpiration: Duration;
```

- *Type:* aws-cdk-lib.Duration
- *Default:* Duration.days(30)

---

##### `tableName`<sup>Optional</sup> <a name="tableName" id="@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.tableName"></a>

```typescript
public readonly tableName: string;
```

- *Type:* string
- *Default:* 'app_logs'

---

##### `workgroupName`<sup>Optional</sup> <a name="workgroupName" id="@jaggr2/cdk-log-to-s3.LogAnalyticsFromExtensionOptions.property.workgroupName"></a>

```typescript
public readonly workgroupName: string;
```

- *Type:* string
- *Default:* derived from the database name

---

### LogAnalyticsProps <a name="LogAnalyticsProps" id="@jaggr2/cdk-log-to-s3.LogAnalyticsProps"></a>

#### Initializer <a name="Initializer" id="@jaggr2/cdk-log-to-s3.LogAnalyticsProps.Initializer"></a>

```typescript
import { LogAnalyticsProps } from '@jaggr2/cdk-log-to-s3'

const logAnalyticsProps: LogAnalyticsProps = { ... }
```

#### Properties <a name="Properties" id="Properties"></a>

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.databaseName">databaseName</a></code> | <code>string</code> | Glue database name. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.logsBucket">logsBucket</a></code> | <code>aws-cdk-lib.aws_s3.IBucket</code> | Bucket the extension writes to. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.createDatabase">createDatabase</a></code> | <code>boolean</code> | Create the Glue database. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.createWorkgroup">createWorkgroup</a></code> | <code>boolean</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.keyPrefix">keyPrefix</a></code> | <code>string</code> | Must equal LogToS3ExtensionProps.keyPrefix. A mismatch produces a table that returns no rows, with no error from Athena or CloudFormation. Use LogAnalytics.fromExtension() to make that impossible. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.projectionWindow">projectionWindow</a></code> | <code>aws-cdk-lib.Duration</code> | How far back the partition projection reaches, as a sliding window ending at today. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.removalPolicy">removalPolicy</a></code> | <code>aws-cdk-lib.RemovalPolicy</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.resultsBucket">resultsBucket</a></code> | <code>aws-cdk-lib.aws_s3.IBucket</code> | Bucket for Athena query results. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.resultsExpiration">resultsExpiration</a></code> | <code>aws-cdk-lib.Duration</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.tableName">tableName</a></code> | <code>string</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.workgroupName">workgroupName</a></code> | <code>string</code> | *No description.* |

---

##### `databaseName`<sup>Required</sup> <a name="databaseName" id="@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.databaseName"></a>

```typescript
public readonly databaseName: string;
```

- *Type:* string

Glue database name.

Lowercase, digits and underscores only.

---

##### `logsBucket`<sup>Required</sup> <a name="logsBucket" id="@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.logsBucket"></a>

```typescript
public readonly logsBucket: IBucket;
```

- *Type:* aws-cdk-lib.aws_s3.IBucket

Bucket the extension writes to.

---

##### `createDatabase`<sup>Optional</sup> <a name="createDatabase" id="@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.createDatabase"></a>

```typescript
public readonly createDatabase: boolean;
```

- *Type:* boolean
- *Default:* true

Create the Glue database.

Set false when it already exists - notably when
migrating from a setup that created it imperatively, since
AWS::Glue::Database fails if the database is already there.

---

##### `createWorkgroup`<sup>Optional</sup> <a name="createWorkgroup" id="@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.createWorkgroup"></a>

```typescript
public readonly createWorkgroup: boolean;
```

- *Type:* boolean
- *Default:* true

---

##### `keyPrefix`<sup>Optional</sup> <a name="keyPrefix" id="@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.keyPrefix"></a>

```typescript
public readonly keyPrefix: string;
```

- *Type:* string
- *Default:* 'logs/'

Must equal LogToS3ExtensionProps.keyPrefix. A mismatch produces a table that returns no rows, with no error from Athena or CloudFormation. Use LogAnalytics.fromExtension() to make that impossible.

---

##### `projectionWindow`<sup>Optional</sup> <a name="projectionWindow" id="@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.projectionWindow"></a>

```typescript
public readonly projectionWindow: Duration;
```

- *Type:* aws-cdk-lib.Duration
- *Default:* Duration.days(730)

How far back the partition projection reaches, as a sliding window ending at today.

Projection enumerates every value in the window, so this is the partition
count: one per day. Because the window slides it stays bounded instead of
growing forever.

It must be wider than how long you keep the data. Objects older than the
window are still in S3 but Athena cannot generate a partition for them, so
they become unqueryable - silently, with no error. The LogBucket default
expires objects after 180 days, well inside this default.

---

##### `removalPolicy`<sup>Optional</sup> <a name="removalPolicy" id="@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.removalPolicy"></a>

```typescript
public readonly removalPolicy: RemovalPolicy;
```

- *Type:* aws-cdk-lib.RemovalPolicy
- *Default:* RemovalPolicy.RETAIN

---

##### `resultsBucket`<sup>Optional</sup> <a name="resultsBucket" id="@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.resultsBucket"></a>

```typescript
public readonly resultsBucket: IBucket;
```

- *Type:* aws-cdk-lib.aws_s3.IBucket
- *Default:* a bucket is created with a 30 day expiry

Bucket for Athena query results.

---

##### `resultsExpiration`<sup>Optional</sup> <a name="resultsExpiration" id="@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.resultsExpiration"></a>

```typescript
public readonly resultsExpiration: Duration;
```

- *Type:* aws-cdk-lib.Duration
- *Default:* Duration.days(30)

---

##### `tableName`<sup>Optional</sup> <a name="tableName" id="@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.tableName"></a>

```typescript
public readonly tableName: string;
```

- *Type:* string
- *Default:* 'app_logs'

---

##### `workgroupName`<sup>Optional</sup> <a name="workgroupName" id="@jaggr2/cdk-log-to-s3.LogAnalyticsProps.property.workgroupName"></a>

```typescript
public readonly workgroupName: string;
```

- *Type:* string
- *Default:* derived from the database name

---

### LogBucketProps <a name="LogBucketProps" id="@jaggr2/cdk-log-to-s3.LogBucketProps"></a>

#### Initializer <a name="Initializer" id="@jaggr2/cdk-log-to-s3.LogBucketProps.Initializer"></a>

```typescript
import { LogBucketProps } from '@jaggr2/cdk-log-to-s3'

const logBucketProps: LogBucketProps = { ... }
```

#### Properties <a name="Properties" id="Properties"></a>

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.accessControl">accessControl</a></code> | <code>aws-cdk-lib.aws_s3.BucketAccessControl</code> | Specifies a canned ACL that grants predefined permissions to the bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.autoDeleteObjects">autoDeleteObjects</a></code> | <code>boolean</code> | Whether all objects should be automatically deleted when the bucket is removed from the stack or when the stack is deleted. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.blockPublicAccess">blockPublicAccess</a></code> | <code>aws-cdk-lib.aws_s3.BlockPublicAccess</code> | The block public access configuration of this bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.bucketKeyEnabled">bucketKeyEnabled</a></code> | <code>boolean</code> | Whether Amazon S3 should use its own intermediary key to generate data keys. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.bucketName">bucketName</a></code> | <code>string</code> | Physical name of this bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.cors">cors</a></code> | <code>aws-cdk-lib.aws_s3.CorsRule[]</code> | The CORS configuration of this bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.encryption">encryption</a></code> | <code>aws-cdk-lib.aws_s3.BucketEncryption</code> | The kind of server-side encryption to apply to this bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.encryptionKey">encryptionKey</a></code> | <code>aws-cdk-lib.aws_kms.IKey</code> | External KMS key to use for bucket encryption. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.enforceSSL">enforceSSL</a></code> | <code>boolean</code> | Enforces SSL for requests. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.eventBridgeEnabled">eventBridgeEnabled</a></code> | <code>boolean</code> | Whether this bucket should send notifications to Amazon EventBridge or not. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.intelligentTieringConfigurations">intelligentTieringConfigurations</a></code> | <code>aws-cdk-lib.aws_s3.IntelligentTieringConfiguration[]</code> | Inteligent Tiering Configurations. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.inventories">inventories</a></code> | <code>aws-cdk-lib.aws_s3.Inventory[]</code> | The inventory configuration of the bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.lifecycleRules">lifecycleRules</a></code> | <code>aws-cdk-lib.aws_s3.LifecycleRule[]</code> | Rules that define how Amazon S3 manages objects during their lifetime. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.metrics">metrics</a></code> | <code>aws-cdk-lib.aws_s3.BucketMetrics[]</code> | The metrics configuration of this bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.minimumTLSVersion">minimumTLSVersion</a></code> | <code>number</code> | Enforces minimum TLS version for requests. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.notificationsHandlerRole">notificationsHandlerRole</a></code> | <code>aws-cdk-lib.aws_iam.IRole</code> | The role to be used by the notifications handler. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.objectLockDefaultRetention">objectLockDefaultRetention</a></code> | <code>aws-cdk-lib.aws_s3.ObjectLockRetention</code> | The default retention mode and rules for S3 Object Lock. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.objectLockEnabled">objectLockEnabled</a></code> | <code>boolean</code> | Enable object lock on the bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.objectOwnership">objectOwnership</a></code> | <code>aws-cdk-lib.aws_s3.ObjectOwnership</code> | The objectOwnership of the bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.publicReadAccess">publicReadAccess</a></code> | <code>boolean</code> | Grants public read access to all objects in the bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.removalPolicy">removalPolicy</a></code> | <code>aws-cdk-lib.RemovalPolicy</code> | Policy to apply when the bucket is removed from this stack. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.serverAccessLogsBucket">serverAccessLogsBucket</a></code> | <code>aws-cdk-lib.aws_s3.IBucket</code> | Destination bucket for the server access logs. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.serverAccessLogsPrefix">serverAccessLogsPrefix</a></code> | <code>string</code> | Optional log file prefix to use for the bucket's access logs. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.targetObjectKeyFormat">targetObjectKeyFormat</a></code> | <code>aws-cdk-lib.aws_s3.TargetObjectKeyFormat</code> | Optional key format for log objects. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.transferAcceleration">transferAcceleration</a></code> | <code>boolean</code> | Whether this bucket should have transfer acceleration turned on or not. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.versioned">versioned</a></code> | <code>boolean</code> | Whether this bucket should have versioning turned on or not. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.websiteErrorDocument">websiteErrorDocument</a></code> | <code>string</code> | The name of the error document (e.g. "404.html") for the website. `websiteIndexDocument` must also be set if this is set. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.websiteIndexDocument">websiteIndexDocument</a></code> | <code>string</code> | The name of the index document (e.g. "index.html") for the website. Enables static website hosting for this bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.websiteRedirect">websiteRedirect</a></code> | <code>aws-cdk-lib.aws_s3.RedirectTarget</code> | Specifies the redirect behavior of all requests to a website endpoint of a bucket. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.websiteRoutingRules">websiteRoutingRules</a></code> | <code>aws-cdk-lib.aws_s3.RoutingRule[]</code> | Rules that define when a redirect is applied and the redirect behavior. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.abortIncompleteUploadsAfter">abortIncompleteUploadsAfter</a></code> | <code>aws-cdk-lib.Duration</code> | Abort incomplete multipart uploads after this long. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.expireAfter">expireAfter</a></code> | <code>aws-cdk-lib.Duration</code> | Delete objects after this long. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.glacierAfter">glacierAfter</a></code> | <code>aws-cdk-lib.Duration</code> | Transition objects to Glacier after this long. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.infrequentAccessAfter">infrequentAccessAfter</a></code> | <code>aws-cdk-lib.Duration</code> | Transition objects to Infrequent Access after this long. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogBucketProps.property.keyPrefix">keyPrefix</a></code> | <code>string</code> | Key prefix the extension writes under. |

---

##### `accessControl`<sup>Optional</sup> <a name="accessControl" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.accessControl"></a>

```typescript
public readonly accessControl: BucketAccessControl;
```

- *Type:* aws-cdk-lib.aws_s3.BucketAccessControl
- *Default:* BucketAccessControl.PRIVATE

Specifies a canned ACL that grants predefined permissions to the bucket.

---

##### `autoDeleteObjects`<sup>Optional</sup> <a name="autoDeleteObjects" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.autoDeleteObjects"></a>

```typescript
public readonly autoDeleteObjects: boolean;
```

- *Type:* boolean
- *Default:* false

Whether all objects should be automatically deleted when the bucket is removed from the stack or when the stack is deleted.

Requires the `removalPolicy` to be set to `RemovalPolicy.DESTROY`.

**Warning** if you have deployed a bucket with `autoDeleteObjects: true`,
switching this to `false` in a CDK version *before* `1.126.0` will lead to
all objects in the bucket being deleted. Be sure to update your bucket resources
by deploying with CDK version `1.126.0` or later **before** switching this value to `false`.

Setting `autoDeleteObjects` to true on a bucket will add `s3:PutBucketPolicy` to the
bucket policy. This is because during bucket deletion, the custom resource provider
needs to update the bucket policy by adding a deny policy for `s3:PutObject` to
prevent race conditions with external bucket writers.

---

##### `blockPublicAccess`<sup>Optional</sup> <a name="blockPublicAccess" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.blockPublicAccess"></a>

```typescript
public readonly blockPublicAccess: BlockPublicAccess;
```

- *Type:* aws-cdk-lib.aws_s3.BlockPublicAccess
- *Default:* CloudFormation defaults will apply. New buckets and objects don't allow public access, but users can modify bucket policies or object permissions to allow public access

The block public access configuration of this bucket.

> [https://docs.aws.amazon.com/AmazonS3/latest/dev/access-control-block-public-access.html](https://docs.aws.amazon.com/AmazonS3/latest/dev/access-control-block-public-access.html)

---

##### `bucketKeyEnabled`<sup>Optional</sup> <a name="bucketKeyEnabled" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.bucketKeyEnabled"></a>

```typescript
public readonly bucketKeyEnabled: boolean;
```

- *Type:* boolean
- *Default:* false

Whether Amazon S3 should use its own intermediary key to generate data keys.

Only relevant when using KMS for encryption.

- If not enabled, every object GET and PUT will cause an API call to KMS (with the
  attendant cost implications of that).
- If enabled, S3 will use its own time-limited key instead.

Only relevant, when Encryption is not set to `BucketEncryption.UNENCRYPTED`.

---

##### `bucketName`<sup>Optional</sup> <a name="bucketName" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.bucketName"></a>

```typescript
public readonly bucketName: string;
```

- *Type:* string
- *Default:* Assigned by CloudFormation (recommended).

Physical name of this bucket.

---

##### `cors`<sup>Optional</sup> <a name="cors" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.cors"></a>

```typescript
public readonly cors: CorsRule[];
```

- *Type:* aws-cdk-lib.aws_s3.CorsRule[]
- *Default:* No CORS configuration.

The CORS configuration of this bucket.

> [https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-cors.html](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-cors.html)

---

##### `encryption`<sup>Optional</sup> <a name="encryption" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.encryption"></a>

```typescript
public readonly encryption: BucketEncryption;
```

- *Type:* aws-cdk-lib.aws_s3.BucketEncryption
- *Default:* `KMS` if `encryptionKey` is specified, or `UNENCRYPTED` otherwise. But if `UNENCRYPTED` is specified, the bucket will be encrypted as `S3_MANAGED` automatically.

The kind of server-side encryption to apply to this bucket.

If you choose KMS, you can specify a KMS key via `encryptionKey`. If
encryption key is not specified, a key will automatically be created.

---

##### `encryptionKey`<sup>Optional</sup> <a name="encryptionKey" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.encryptionKey"></a>

```typescript
public readonly encryptionKey: IKey;
```

- *Type:* aws-cdk-lib.aws_kms.IKey
- *Default:* If `encryption` is set to `KMS` and this property is undefined, a new KMS key will be created and associated with this bucket.

External KMS key to use for bucket encryption.

The `encryption` property must be either not specified or set to `KMS` or `DSSE`.
An error will be emitted if `encryption` is set to `UNENCRYPTED` or `S3_MANAGED`.

---

##### `enforceSSL`<sup>Optional</sup> <a name="enforceSSL" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.enforceSSL"></a>

```typescript
public readonly enforceSSL: boolean;
```

- *Type:* boolean
- *Default:* false

Enforces SSL for requests.

S3.5 of the AWS Foundational Security Best Practices Regarding S3.

> [https://docs.aws.amazon.com/config/latest/developerguide/s3-bucket-ssl-requests-only.html](https://docs.aws.amazon.com/config/latest/developerguide/s3-bucket-ssl-requests-only.html)

---

##### `eventBridgeEnabled`<sup>Optional</sup> <a name="eventBridgeEnabled" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.eventBridgeEnabled"></a>

```typescript
public readonly eventBridgeEnabled: boolean;
```

- *Type:* boolean
- *Default:* false

Whether this bucket should send notifications to Amazon EventBridge or not.

---

##### `intelligentTieringConfigurations`<sup>Optional</sup> <a name="intelligentTieringConfigurations" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.intelligentTieringConfigurations"></a>

```typescript
public readonly intelligentTieringConfigurations: IntelligentTieringConfiguration[];
```

- *Type:* aws-cdk-lib.aws_s3.IntelligentTieringConfiguration[]
- *Default:* No Intelligent Tiiering Configurations.

Inteligent Tiering Configurations.

> [https://docs.aws.amazon.com/AmazonS3/latest/userguide/intelligent-tiering.html](https://docs.aws.amazon.com/AmazonS3/latest/userguide/intelligent-tiering.html)

---

##### `inventories`<sup>Optional</sup> <a name="inventories" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.inventories"></a>

```typescript
public readonly inventories: Inventory[];
```

- *Type:* aws-cdk-lib.aws_s3.Inventory[]
- *Default:* No inventory configuration

The inventory configuration of the bucket.

> [https://docs.aws.amazon.com/AmazonS3/latest/dev/storage-inventory.html](https://docs.aws.amazon.com/AmazonS3/latest/dev/storage-inventory.html)

---

##### `lifecycleRules`<sup>Optional</sup> <a name="lifecycleRules" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.lifecycleRules"></a>

```typescript
public readonly lifecycleRules: LifecycleRule[];
```

- *Type:* aws-cdk-lib.aws_s3.LifecycleRule[]
- *Default:* No lifecycle rules.

Rules that define how Amazon S3 manages objects during their lifetime.

---

##### `metrics`<sup>Optional</sup> <a name="metrics" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.metrics"></a>

```typescript
public readonly metrics: BucketMetrics[];
```

- *Type:* aws-cdk-lib.aws_s3.BucketMetrics[]
- *Default:* No metrics configuration.

The metrics configuration of this bucket.

> [https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-metricsconfiguration.html](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-s3-bucket-metricsconfiguration.html)

---

##### `minimumTLSVersion`<sup>Optional</sup> <a name="minimumTLSVersion" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.minimumTLSVersion"></a>

```typescript
public readonly minimumTLSVersion: number;
```

- *Type:* number
- *Default:* No minimum TLS version is enforced.

Enforces minimum TLS version for requests.

Requires `enforceSSL` to be enabled.

> [https://docs.aws.amazon.com/AmazonS3/latest/userguide/amazon-s3-policy-keys.html#example-object-tls-version](https://docs.aws.amazon.com/AmazonS3/latest/userguide/amazon-s3-policy-keys.html#example-object-tls-version)

---

##### `notificationsHandlerRole`<sup>Optional</sup> <a name="notificationsHandlerRole" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.notificationsHandlerRole"></a>

```typescript
public readonly notificationsHandlerRole: IRole;
```

- *Type:* aws-cdk-lib.aws_iam.IRole
- *Default:* a new role will be created.

The role to be used by the notifications handler.

---

##### `objectLockDefaultRetention`<sup>Optional</sup> <a name="objectLockDefaultRetention" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.objectLockDefaultRetention"></a>

```typescript
public readonly objectLockDefaultRetention: ObjectLockRetention;
```

- *Type:* aws-cdk-lib.aws_s3.ObjectLockRetention
- *Default:* no default retention period

The default retention mode and rules for S3 Object Lock.

Default retention can be configured after a bucket is created if the bucket already
has object lock enabled. Enabling object lock for existing buckets is not supported.

> [https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lock-overview.html#object-lock-bucket-config-enable](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lock-overview.html#object-lock-bucket-config-enable)

---

##### `objectLockEnabled`<sup>Optional</sup> <a name="objectLockEnabled" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.objectLockEnabled"></a>

```typescript
public readonly objectLockEnabled: boolean;
```

- *Type:* boolean
- *Default:* false, unless objectLockDefaultRetention is set (then, true)

Enable object lock on the bucket.

Enabling object lock for existing buckets is not supported. Object lock must be
enabled when the bucket is created.

> [https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lock-overview.html#object-lock-bucket-config-enable](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lock-overview.html#object-lock-bucket-config-enable)

---

##### `objectOwnership`<sup>Optional</sup> <a name="objectOwnership" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.objectOwnership"></a>

```typescript
public readonly objectOwnership: ObjectOwnership;
```

- *Type:* aws-cdk-lib.aws_s3.ObjectOwnership
- *Default:* No ObjectOwnership configuration, uploading account will own the object.

The objectOwnership of the bucket.

> [https://docs.aws.amazon.com/AmazonS3/latest/dev/about-object-ownership.html](https://docs.aws.amazon.com/AmazonS3/latest/dev/about-object-ownership.html)

---

##### `publicReadAccess`<sup>Optional</sup> <a name="publicReadAccess" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.publicReadAccess"></a>

```typescript
public readonly publicReadAccess: boolean;
```

- *Type:* boolean
- *Default:* false

Grants public read access to all objects in the bucket.

Similar to calling `bucket.grantPublicAccess()`

---

##### `removalPolicy`<sup>Optional</sup> <a name="removalPolicy" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.removalPolicy"></a>

```typescript
public readonly removalPolicy: RemovalPolicy;
```

- *Type:* aws-cdk-lib.RemovalPolicy
- *Default:* The bucket will be orphaned.

Policy to apply when the bucket is removed from this stack.

---

##### `serverAccessLogsBucket`<sup>Optional</sup> <a name="serverAccessLogsBucket" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.serverAccessLogsBucket"></a>

```typescript
public readonly serverAccessLogsBucket: IBucket;
```

- *Type:* aws-cdk-lib.aws_s3.IBucket
- *Default:* If "serverAccessLogsPrefix" undefined - access logs disabled, otherwise - log to current bucket.

Destination bucket for the server access logs.

---

##### `serverAccessLogsPrefix`<sup>Optional</sup> <a name="serverAccessLogsPrefix" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.serverAccessLogsPrefix"></a>

```typescript
public readonly serverAccessLogsPrefix: string;
```

- *Type:* string
- *Default:* No log file prefix

Optional log file prefix to use for the bucket's access logs.

If defined without "serverAccessLogsBucket", enables access logs to current bucket with this prefix.

---

##### `targetObjectKeyFormat`<sup>Optional</sup> <a name="targetObjectKeyFormat" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.targetObjectKeyFormat"></a>

```typescript
public readonly targetObjectKeyFormat: TargetObjectKeyFormat;
```

- *Type:* aws-cdk-lib.aws_s3.TargetObjectKeyFormat
- *Default:* the default key format is: [DestinationPrefix][YYYY]-[MM]-[DD]-[hh]-[mm]-[ss]-[UniqueString]

Optional key format for log objects.

---

##### `transferAcceleration`<sup>Optional</sup> <a name="transferAcceleration" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.transferAcceleration"></a>

```typescript
public readonly transferAcceleration: boolean;
```

- *Type:* boolean
- *Default:* false

Whether this bucket should have transfer acceleration turned on or not.

---

##### `versioned`<sup>Optional</sup> <a name="versioned" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.versioned"></a>

```typescript
public readonly versioned: boolean;
```

- *Type:* boolean
- *Default:* false (unless object lock is enabled, then true)

Whether this bucket should have versioning turned on or not.

---

##### `websiteErrorDocument`<sup>Optional</sup> <a name="websiteErrorDocument" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.websiteErrorDocument"></a>

```typescript
public readonly websiteErrorDocument: string;
```

- *Type:* string
- *Default:* No error document.

The name of the error document (e.g. "404.html") for the website. `websiteIndexDocument` must also be set if this is set.

---

##### `websiteIndexDocument`<sup>Optional</sup> <a name="websiteIndexDocument" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.websiteIndexDocument"></a>

```typescript
public readonly websiteIndexDocument: string;
```

- *Type:* string
- *Default:* No index document.

The name of the index document (e.g. "index.html") for the website. Enables static website hosting for this bucket.

---

##### `websiteRedirect`<sup>Optional</sup> <a name="websiteRedirect" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.websiteRedirect"></a>

```typescript
public readonly websiteRedirect: RedirectTarget;
```

- *Type:* aws-cdk-lib.aws_s3.RedirectTarget
- *Default:* No redirection.

Specifies the redirect behavior of all requests to a website endpoint of a bucket.

If you specify this property, you can't specify "websiteIndexDocument", "websiteErrorDocument" nor , "websiteRoutingRules".

---

##### `websiteRoutingRules`<sup>Optional</sup> <a name="websiteRoutingRules" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.websiteRoutingRules"></a>

```typescript
public readonly websiteRoutingRules: RoutingRule[];
```

- *Type:* aws-cdk-lib.aws_s3.RoutingRule[]
- *Default:* No redirection rules.

Rules that define when a redirect is applied and the redirect behavior.

---

##### `abortIncompleteUploadsAfter`<sup>Optional</sup> <a name="abortIncompleteUploadsAfter" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.abortIncompleteUploadsAfter"></a>

```typescript
public readonly abortIncompleteUploadsAfter: Duration;
```

- *Type:* aws-cdk-lib.Duration
- *Default:* Duration.days(7)

Abort incomplete multipart uploads after this long.

---

##### `expireAfter`<sup>Optional</sup> <a name="expireAfter" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.expireAfter"></a>

```typescript
public readonly expireAfter: Duration;
```

- *Type:* aws-cdk-lib.Duration
- *Default:* Duration.days(180)

Delete objects after this long.

---

##### `glacierAfter`<sup>Optional</sup> <a name="glacierAfter" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.glacierAfter"></a>

```typescript
public readonly glacierAfter: Duration;
```

- *Type:* aws-cdk-lib.Duration
- *Default:* Duration.days(90)

Transition objects to Glacier after this long.

---

##### `infrequentAccessAfter`<sup>Optional</sup> <a name="infrequentAccessAfter" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.infrequentAccessAfter"></a>

```typescript
public readonly infrequentAccessAfter: Duration;
```

- *Type:* aws-cdk-lib.Duration
- *Default:* Duration.days(30)

Transition objects to Infrequent Access after this long.

---

##### `keyPrefix`<sup>Optional</sup> <a name="keyPrefix" id="@jaggr2/cdk-log-to-s3.LogBucketProps.property.keyPrefix"></a>

```typescript
public readonly keyPrefix: string;
```

- *Type:* string
- *Default:* 'logs/'

Key prefix the extension writes under.

Recorded on the construct so a
LogAnalytics table can be pointed at the same place without repeating it.

---

### LogCompactionFromExtensionOptions <a name="LogCompactionFromExtensionOptions" id="@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions"></a>

Options for LogCompaction.fromExtension.

#### Initializer <a name="Initializer" id="@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.Initializer"></a>

```typescript
import { LogCompactionFromExtensionOptions } from '@jaggr2/cdk-log-to-s3'

const logCompactionFromExtensionOptions: LogCompactionFromExtensionOptions = { ... }
```

#### Properties <a name="Properties" id="Properties"></a>

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.compression">compression</a></code> | <code><a href="#@jaggr2/cdk-log-to-s3.LogCompression">LogCompression</a></code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.debug">debug</a></code> | <code>boolean</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.enabled">enabled</a></code> | <code>boolean</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.logRetention">logRetention</a></code> | <code>aws-cdk-lib.aws_logs.RetentionDays</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.lookback">lookback</a></code> | <code>aws-cdk-lib.Duration</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.maxBytesPerRun">maxBytesPerRun</a></code> | <code>aws-cdk-lib.Size</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.maxFilesPerRun">maxFilesPerRun</a></code> | <code>number</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.memorySize">memorySize</a></code> | <code>aws-cdk-lib.Size</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.minFilesPerPartition">minFilesPerPartition</a></code> | <code>number</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.removalPolicy">removalPolicy</a></code> | <code>aws-cdk-lib.RemovalPolicy</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.schedule">schedule</a></code> | <code>aws-cdk-lib.aws_events.Schedule</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.timeout">timeout</a></code> | <code>aws-cdk-lib.Duration</code> | *No description.* |

---

##### `compression`<sup>Optional</sup> <a name="compression" id="@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.compression"></a>

```typescript
public readonly compression: LogCompression;
```

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogCompression">LogCompression</a>
- *Default:* LogCompression.SNAPPY

---

##### `debug`<sup>Optional</sup> <a name="debug" id="@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.debug"></a>

```typescript
public readonly debug: boolean;
```

- *Type:* boolean
- *Default:* false

---

##### `enabled`<sup>Optional</sup> <a name="enabled" id="@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.enabled"></a>

```typescript
public readonly enabled: boolean;
```

- *Type:* boolean
- *Default:* true

---

##### `logRetention`<sup>Optional</sup> <a name="logRetention" id="@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.logRetention"></a>

```typescript
public readonly logRetention: RetentionDays;
```

- *Type:* aws-cdk-lib.aws_logs.RetentionDays
- *Default:* logs.RetentionDays.ONE_MONTH

---

##### `lookback`<sup>Optional</sup> <a name="lookback" id="@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.lookback"></a>

```typescript
public readonly lookback: Duration;
```

- *Type:* aws-cdk-lib.Duration
- *Default:* Duration.days(7)

---

##### `maxBytesPerRun`<sup>Optional</sup> <a name="maxBytesPerRun" id="@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.maxBytesPerRun"></a>

```typescript
public readonly maxBytesPerRun: Size;
```

- *Type:* aws-cdk-lib.Size
- *Default:* Size.mebibytes(256)

---

##### `maxFilesPerRun`<sup>Optional</sup> <a name="maxFilesPerRun" id="@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.maxFilesPerRun"></a>

```typescript
public readonly maxFilesPerRun: number;
```

- *Type:* number
- *Default:* 2000

---

##### `memorySize`<sup>Optional</sup> <a name="memorySize" id="@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.memorySize"></a>

```typescript
public readonly memorySize: Size;
```

- *Type:* aws-cdk-lib.Size
- *Default:* Size.mebibytes(1024)

---

##### `minFilesPerPartition`<sup>Optional</sup> <a name="minFilesPerPartition" id="@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.minFilesPerPartition"></a>

```typescript
public readonly minFilesPerPartition: number;
```

- *Type:* number
- *Default:* 8

---

##### `removalPolicy`<sup>Optional</sup> <a name="removalPolicy" id="@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.removalPolicy"></a>

```typescript
public readonly removalPolicy: RemovalPolicy;
```

- *Type:* aws-cdk-lib.RemovalPolicy
- *Default:* RemovalPolicy.DESTROY

---

##### `schedule`<sup>Optional</sup> <a name="schedule" id="@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.schedule"></a>

```typescript
public readonly schedule: Schedule;
```

- *Type:* aws-cdk-lib.aws_events.Schedule
- *Default:* every day at 03:00 UTC

---

##### `timeout`<sup>Optional</sup> <a name="timeout" id="@jaggr2/cdk-log-to-s3.LogCompactionFromExtensionOptions.property.timeout"></a>

```typescript
public readonly timeout: Duration;
```

- *Type:* aws-cdk-lib.Duration
- *Default:* Duration.minutes(5)

---

### LogCompactionProps <a name="LogCompactionProps" id="@jaggr2/cdk-log-to-s3.LogCompactionProps"></a>

#### Initializer <a name="Initializer" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.Initializer"></a>

```typescript
import { LogCompactionProps } from '@jaggr2/cdk-log-to-s3'

const logCompactionProps: LogCompactionProps = { ... }
```

#### Properties <a name="Properties" id="Properties"></a>

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps.property.logsBucket">logsBucket</a></code> | <code>aws-cdk-lib.aws_s3.IBucket</code> | Bucket the extension writes to. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps.property.architecture">architecture</a></code> | <code>aws-cdk-lib.aws_lambda.Architecture</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps.property.compression">compression</a></code> | <code><a href="#@jaggr2/cdk-log-to-s3.LogCompression">LogCompression</a></code> | Codec for the merged output. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps.property.debug">debug</a></code> | <code>boolean</code> | Report every partition considered, not only the ones compacted. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps.property.enabled">enabled</a></code> | <code>boolean</code> | Create the schedule. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps.property.keyPrefix">keyPrefix</a></code> | <code>string</code> | Must equal LogToS3ExtensionProps.keyPrefix, or the job will look in the wrong place and quietly find nothing to do. Use LogCompaction.fromExtension() to make that impossible. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps.property.logRetention">logRetention</a></code> | <code>aws-cdk-lib.aws_logs.RetentionDays</code> | Retention for the job's own CloudWatch logs. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps.property.lookback">lookback</a></code> | <code>aws-cdk-lib.Duration</code> | How many closed days each run considers, most recent first. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps.property.maxBytesPerRun">maxBytesPerRun</a></code> | <code>aws-cdk-lib.Size</code> | Cap on bytes read per partition per run. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps.property.maxFilesPerRun">maxFilesPerRun</a></code> | <code>number</code> | Cap on files merged per partition per run. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps.property.memorySize">memorySize</a></code> | <code>aws-cdk-lib.Size</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps.property.minFilesPerPartition">minFilesPerPartition</a></code> | <code>number</code> | Leave a partition alone until it holds at least this many files. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps.property.removalPolicy">removalPolicy</a></code> | <code>aws-cdk-lib.RemovalPolicy</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps.property.schedule">schedule</a></code> | <code>aws-cdk-lib.aws_events.Schedule</code> | When to run. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompactionProps.property.timeout">timeout</a></code> | <code>aws-cdk-lib.Duration</code> | *No description.* |

---

##### `logsBucket`<sup>Required</sup> <a name="logsBucket" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.property.logsBucket"></a>

```typescript
public readonly logsBucket: IBucket;
```

- *Type:* aws-cdk-lib.aws_s3.IBucket

Bucket the extension writes to.

---

##### `architecture`<sup>Optional</sup> <a name="architecture" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.property.architecture"></a>

```typescript
public readonly architecture: Architecture;
```

- *Type:* aws-cdk-lib.aws_lambda.Architecture
- *Default:* lambda.Architecture.ARM_64

---

##### `compression`<sup>Optional</sup> <a name="compression" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.property.compression"></a>

```typescript
public readonly compression: LogCompression;
```

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogCompression">LogCompression</a>
- *Default:* LogCompression.SNAPPY

Codec for the merged output.

Match the extension unless you want the
archive stored more densely than it was written.

---

##### `debug`<sup>Optional</sup> <a name="debug" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.property.debug"></a>

```typescript
public readonly debug: boolean;
```

- *Type:* boolean
- *Default:* false

Report every partition considered, not only the ones compacted.

---

##### `enabled`<sup>Optional</sup> <a name="enabled" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.property.enabled"></a>

```typescript
public readonly enabled: boolean;
```

- *Type:* boolean
- *Default:* true

Create the schedule.

Set false to deploy the function but trigger it
yourself.

---

##### `keyPrefix`<sup>Optional</sup> <a name="keyPrefix" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.property.keyPrefix"></a>

```typescript
public readonly keyPrefix: string;
```

- *Type:* string
- *Default:* 'logs/'

Must equal LogToS3ExtensionProps.keyPrefix, or the job will look in the wrong place and quietly find nothing to do. Use LogCompaction.fromExtension() to make that impossible.

---

##### `logRetention`<sup>Optional</sup> <a name="logRetention" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.property.logRetention"></a>

```typescript
public readonly logRetention: RetentionDays;
```

- *Type:* aws-cdk-lib.aws_logs.RetentionDays
- *Default:* logs.RetentionDays.ONE_MONTH

Retention for the job's own CloudWatch logs.

---

##### `lookback`<sup>Optional</sup> <a name="lookback" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.property.lookback"></a>

```typescript
public readonly lookback: Duration;
```

- *Type:* aws-cdk-lib.Duration
- *Default:* Duration.days(7)

How many closed days each run considers, most recent first.

More than one
so a failed or skipped run catches up by itself.

---

##### `maxBytesPerRun`<sup>Optional</sup> <a name="maxBytesPerRun" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.property.maxBytesPerRun"></a>

```typescript
public readonly maxBytesPerRun: Size;
```

- *Type:* aws-cdk-lib.Size
- *Default:* Size.mebibytes(256)

Cap on bytes read per partition per run.

Keep it comfortably below
`memorySize`: rows are held in memory while merging.

---

##### `maxFilesPerRun`<sup>Optional</sup> <a name="maxFilesPerRun" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.property.maxFilesPerRun"></a>

```typescript
public readonly maxFilesPerRun: number;
```

- *Type:* number
- *Default:* 2000

Cap on files merged per partition per run.

Whatever is left over is picked
up next run, and every run still reduces the file count.

---

##### `memorySize`<sup>Optional</sup> <a name="memorySize" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.property.memorySize"></a>

```typescript
public readonly memorySize: Size;
```

- *Type:* aws-cdk-lib.Size
- *Default:* Size.mebibytes(1024)

---

##### `minFilesPerPartition`<sup>Optional</sup> <a name="minFilesPerPartition" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.property.minFilesPerPartition"></a>

```typescript
public readonly minFilesPerPartition: number;
```

- *Type:* number
- *Default:* 8

Leave a partition alone until it holds at least this many files.

Below it
the read and rewrite costs more than the scan it saves.

---

##### `removalPolicy`<sup>Optional</sup> <a name="removalPolicy" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.property.removalPolicy"></a>

```typescript
public readonly removalPolicy: RemovalPolicy;
```

- *Type:* aws-cdk-lib.RemovalPolicy
- *Default:* RemovalPolicy.DESTROY

---

##### `schedule`<sup>Optional</sup> <a name="schedule" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.property.schedule"></a>

```typescript
public readonly schedule: Schedule;
```

- *Type:* aws-cdk-lib.aws_events.Schedule
- *Default:* every day at 03:00 UTC

When to run.

Compaction only touches closed days, so anything daily works;
off-peak is polite but not required.

---

##### `timeout`<sup>Optional</sup> <a name="timeout" id="@jaggr2/cdk-log-to-s3.LogCompactionProps.property.timeout"></a>

```typescript
public readonly timeout: Duration;
```

- *Type:* aws-cdk-lib.Duration
- *Default:* Duration.minutes(5)

---

### LogToS3AttachOptions <a name="LogToS3AttachOptions" id="@jaggr2/cdk-log-to-s3.LogToS3AttachOptions"></a>

#### Initializer <a name="Initializer" id="@jaggr2/cdk-log-to-s3.LogToS3AttachOptions.Initializer"></a>

```typescript
import { LogToS3AttachOptions } from '@jaggr2/cdk-log-to-s3'

const logToS3AttachOptions: LogToS3AttachOptions = { ... }
```

#### Properties <a name="Properties" id="Properties"></a>

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3AttachOptions.property.grantWrite">grantWrite</a></code> | <code>boolean</code> | Grant the function permission to write to the logs bucket, scoped to the key prefix. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3AttachOptions.property.logLevel">logLevel</a></code> | <code><a href="#@jaggr2/cdk-log-to-s3.LogLevel">LogLevel</a></code> | Override the extension-wide level for this one function. |

---

##### `grantWrite`<sup>Optional</sup> <a name="grantWrite" id="@jaggr2/cdk-log-to-s3.LogToS3AttachOptions.property.grantWrite"></a>

```typescript
public readonly grantWrite: boolean;
```

- *Type:* boolean
- *Default:* true

Grant the function permission to write to the logs bucket, scoped to the key prefix.

Set false only if the role already has equivalent access.

---

##### `logLevel`<sup>Optional</sup> <a name="logLevel" id="@jaggr2/cdk-log-to-s3.LogToS3AttachOptions.property.logLevel"></a>

```typescript
public readonly logLevel: LogLevel;
```

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogLevel">LogLevel</a>
- *Default:* the level configured on the extension

Override the extension-wide level for this one function.

---

### LogToS3ExtensionAttributes <a name="LogToS3ExtensionAttributes" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes"></a>

Attributes needed to use a layer that was created elsewhere.

#### Initializer <a name="Initializer" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes.Initializer"></a>

```typescript
import { LogToS3ExtensionAttributes } from '@jaggr2/cdk-log-to-s3'

const logToS3ExtensionAttributes: LogToS3ExtensionAttributes = { ... }
```

#### Properties <a name="Properties" id="Properties"></a>

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes.property.layerVersionArn">layerVersionArn</a></code> | <code>string</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes.property.logsBucket">logsBucket</a></code> | <code>aws-cdk-lib.aws_s3.IBucket</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes.property.architecture">architecture</a></code> | <code>aws-cdk-lib.aws_lambda.Architecture</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes.property.compression">compression</a></code> | <code><a href="#@jaggr2/cdk-log-to-s3.LogCompression">LogCompression</a></code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes.property.keyPrefix">keyPrefix</a></code> | <code>string</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes.property.logLevel">logLevel</a></code> | <code><a href="#@jaggr2/cdk-log-to-s3.LogLevel">LogLevel</a></code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes.property.telemetryPort">telemetryPort</a></code> | <code>number</code> | *No description.* |

---

##### `layerVersionArn`<sup>Required</sup> <a name="layerVersionArn" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes.property.layerVersionArn"></a>

```typescript
public readonly layerVersionArn: string;
```

- *Type:* string

---

##### `logsBucket`<sup>Required</sup> <a name="logsBucket" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes.property.logsBucket"></a>

```typescript
public readonly logsBucket: IBucket;
```

- *Type:* aws-cdk-lib.aws_s3.IBucket

---

##### `architecture`<sup>Optional</sup> <a name="architecture" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes.property.architecture"></a>

```typescript
public readonly architecture: Architecture;
```

- *Type:* aws-cdk-lib.aws_lambda.Architecture
- *Default:* lambda.Architecture.ARM_64

---

##### `compression`<sup>Optional</sup> <a name="compression" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes.property.compression"></a>

```typescript
public readonly compression: LogCompression;
```

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogCompression">LogCompression</a>
- *Default:* LogCompression.SNAPPY

---

##### `keyPrefix`<sup>Optional</sup> <a name="keyPrefix" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes.property.keyPrefix"></a>

```typescript
public readonly keyPrefix: string;
```

- *Type:* string
- *Default:* 'logs/'

---

##### `logLevel`<sup>Optional</sup> <a name="logLevel" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes.property.logLevel"></a>

```typescript
public readonly logLevel: LogLevel;
```

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogLevel">LogLevel</a>
- *Default:* LogLevel.INFO

---

##### `telemetryPort`<sup>Optional</sup> <a name="telemetryPort" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionAttributes.property.telemetryPort"></a>

```typescript
public readonly telemetryPort: number;
```

- *Type:* number
- *Default:* 2020

---

### LogToS3ExtensionProps <a name="LogToS3ExtensionProps" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps"></a>

#### Initializer <a name="Initializer" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.Initializer"></a>

```typescript
import { LogToS3ExtensionProps } from '@jaggr2/cdk-log-to-s3'

const logToS3ExtensionProps: LogToS3ExtensionProps = { ... }
```

#### Properties <a name="Properties" id="Properties"></a>

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.logsBucket">logsBucket</a></code> | <code>aws-cdk-lib.aws_s3.IBucket</code> | Bucket the extension writes Parquet files to. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.architecture">architecture</a></code> | <code>aws-cdk-lib.aws_lambda.Architecture</code> | Architecture of the functions this extension will be attached to. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.compatibleRuntimes">compatibleRuntimes</a></code> | <code>aws-cdk-lib.aws_lambda.Runtime[]</code> | Runtimes the layer declares compatibility with. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.compression">compression</a></code> | <code><a href="#@jaggr2/cdk-log-to-s3.LogCompression">LogCompression</a></code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.description">description</a></code> | <code>string</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.extensionDebug">extensionDebug</a></code> | <code>boolean</code> | Verbose self-logging from the extension. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.flushInterval">flushInterval</a></code> | <code>aws-cdk-lib.Duration</code> | How often buffered records are written out, independent of size. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.includePlatformReport">includePlatformReport</a></code> | <code>boolean</code> | Emit one row per invocation with duration and memory metrics, derived from the platform.report event. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.keyPrefix">keyPrefix</a></code> | <code>string</code> | Key prefix to write under. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.layerVersionName">layerVersionName</a></code> | <code>string</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.logLevel">logLevel</a></code> | <code><a href="#@jaggr2/cdk-log-to-s3.LogLevel">LogLevel</a></code> | Records below this level are dropped before being buffered. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.maxBufferSize">maxBufferSize</a></code> | <code>aws-cdk-lib.Size</code> | Buffer size that triggers a flush. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.removalPolicy">removalPolicy</a></code> | <code>aws-cdk-lib.RemovalPolicy</code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.telemetryPort">telemetryPort</a></code> | <code>number</code> | Port the extension listens on for the Telemetry API. |

---

##### `logsBucket`<sup>Required</sup> <a name="logsBucket" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.logsBucket"></a>

```typescript
public readonly logsBucket: IBucket;
```

- *Type:* aws-cdk-lib.aws_s3.IBucket

Bucket the extension writes Parquet files to.

Any IBucket works: the
LogBucket from this package, a plain s3.Bucket, or an imported one.

---

##### `architecture`<sup>Optional</sup> <a name="architecture" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.architecture"></a>

```typescript
public readonly architecture: Architecture;
```

- *Type:* aws-cdk-lib.aws_lambda.Architecture
- *Default:* lambda.Architecture.ARM_64

Architecture of the functions this extension will be attached to.

A layer
carries a native binary, so one instance serves one architecture; create a
second instance for functions on the other.

---

##### `compatibleRuntimes`<sup>Optional</sup> <a name="compatibleRuntimes" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.compatibleRuntimes"></a>

```typescript
public readonly compatibleRuntimes: Runtime[];
```

- *Type:* aws-cdk-lib.aws_lambda.Runtime[]
- *Default:* all runtimes

Runtimes the layer declares compatibility with.

Left unset by default: the
extension is a standalone binary and works with every runtime, so pinning
a list only blocks legitimate consumers.

---

##### `compression`<sup>Optional</sup> <a name="compression" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.compression"></a>

```typescript
public readonly compression: LogCompression;
```

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogCompression">LogCompression</a>
- *Default:* LogCompression.SNAPPY

---

##### `description`<sup>Optional</sup> <a name="description" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.description"></a>

```typescript
public readonly description: string;
```

- *Type:* string
- *Default:* a description mentioning the Telemetry API

---

##### `extensionDebug`<sup>Optional</sup> <a name="extensionDebug" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.extensionDebug"></a>

```typescript
public readonly extensionDebug: boolean;
```

- *Type:* boolean
- *Default:* false

Verbose self-logging from the extension.

Off by default: every line it
writes is itself billed CloudWatch ingest on every invocation.

---

##### `flushInterval`<sup>Optional</sup> <a name="flushInterval" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.flushInterval"></a>

```typescript
public readonly flushInterval: Duration;
```

- *Type:* aws-cdk-lib.Duration
- *Default:* Duration.seconds(15)

How often buffered records are written out, independent of size.

Lower
values mean fresher data in Athena and more, smaller S3 objects.

---

##### `includePlatformReport`<sup>Optional</sup> <a name="includePlatformReport" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.includePlatformReport"></a>

```typescript
public readonly includePlatformReport: boolean;
```

- *Type:* boolean
- *Default:* true

Emit one row per invocation with duration and memory metrics, derived from the platform.report event.

---

##### `keyPrefix`<sup>Optional</sup> <a name="keyPrefix" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.keyPrefix"></a>

```typescript
public readonly keyPrefix: string;
```

- *Type:* string
- *Default:* 'logs/'

Key prefix to write under.

Must match LogAnalyticsProps.keyPrefix, or the
table will be empty with no error reported anywhere.

---

##### `layerVersionName`<sup>Optional</sup> <a name="layerVersionName" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.layerVersionName"></a>

```typescript
public readonly layerVersionName: string;
```

- *Type:* string
- *Default:* a name is generated

---

##### `logLevel`<sup>Optional</sup> <a name="logLevel" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.logLevel"></a>

```typescript
public readonly logLevel: LogLevel;
```

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogLevel">LogLevel</a>
- *Default:* LogLevel.INFO

Records below this level are dropped before being buffered.

---

##### `maxBufferSize`<sup>Optional</sup> <a name="maxBufferSize" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.maxBufferSize"></a>

```typescript
public readonly maxBufferSize: Size;
```

- *Type:* aws-cdk-lib.Size
- *Default:* Size.mebibytes(10)

Buffer size that triggers a flush.

This memory is charged to the function.

---

##### `removalPolicy`<sup>Optional</sup> <a name="removalPolicy" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.removalPolicy"></a>

```typescript
public readonly removalPolicy: RemovalPolicy;
```

- *Type:* aws-cdk-lib.RemovalPolicy
- *Default:* RemovalPolicy.DESTROY

---

##### `telemetryPort`<sup>Optional</sup> <a name="telemetryPort" id="@jaggr2/cdk-log-to-s3.LogToS3ExtensionProps.property.telemetryPort"></a>

```typescript
public readonly telemetryPort: number;
```

- *Type:* number
- *Default:* 2020

Port the extension listens on for the Telemetry API.

Change it only if it
collides with something else in the sandbox.

---


## Protocols <a name="Protocols" id="Protocols"></a>

### ILogToS3Extension <a name="ILogToS3Extension" id="@jaggr2/cdk-log-to-s3.ILogToS3Extension"></a>

- *Extends:* constructs.IConstruct

- *Implemented By:* <a href="#@jaggr2/cdk-log-to-s3.LogToS3Extension">LogToS3Extension</a>, <a href="#@jaggr2/cdk-log-to-s3.ILogToS3Extension">ILogToS3Extension</a>

#### Methods <a name="Methods" id="Methods"></a>

| **Name** | **Description** |
| --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.ILogToS3Extension.attachTo">attachTo</a></code> | Add the layer, the environment and the bucket grant to a function. |
| <code><a href="#@jaggr2/cdk-log-to-s3.ILogToS3Extension.grantWriteLogs">grantWriteLogs</a></code> | Grant write access to the log prefix without attaching the layer. |

---

##### `attachTo` <a name="attachTo" id="@jaggr2/cdk-log-to-s3.ILogToS3Extension.attachTo"></a>

```typescript
public attachTo(fn: Function, options?: LogToS3AttachOptions): void
```

Add the layer, the environment and the bucket grant to a function.

Takes a concrete lambda.Function rather than an IFunction because
addLayers() and addEnvironment() only exist on the concrete class.
Functions imported with Function.fromFunctionArn cannot be modified by
CloudFormation anyway.

###### `fn`<sup>Required</sup> <a name="fn" id="@jaggr2/cdk-log-to-s3.ILogToS3Extension.attachTo.parameter.fn"></a>

- *Type:* aws-cdk-lib.aws_lambda.Function

---

###### `options`<sup>Optional</sup> <a name="options" id="@jaggr2/cdk-log-to-s3.ILogToS3Extension.attachTo.parameter.options"></a>

- *Type:* <a href="#@jaggr2/cdk-log-to-s3.LogToS3AttachOptions">LogToS3AttachOptions</a>

---

##### `grantWriteLogs` <a name="grantWriteLogs" id="@jaggr2/cdk-log-to-s3.ILogToS3Extension.grantWriteLogs"></a>

```typescript
public grantWriteLogs(grantee: IGrantable): Grant
```

Grant write access to the log prefix without attaching the layer.

###### `grantee`<sup>Required</sup> <a name="grantee" id="@jaggr2/cdk-log-to-s3.ILogToS3Extension.grantWriteLogs.parameter.grantee"></a>

- *Type:* aws-cdk-lib.aws_iam.IGrantable

---

#### Properties <a name="Properties" id="Properties"></a>

| **Name** | **Type** | **Description** |
| --- | --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.ILogToS3Extension.property.node">node</a></code> | <code>constructs.Node</code> | The tree node. |
| <code><a href="#@jaggr2/cdk-log-to-s3.ILogToS3Extension.property.architecture">architecture</a></code> | <code>aws-cdk-lib.aws_lambda.Architecture</code> | Architecture this layer was built for. |
| <code><a href="#@jaggr2/cdk-log-to-s3.ILogToS3Extension.property.environment">environment</a></code> | <code>{[ key: string ]: string}</code> | Environment variables attachTo() injects. |
| <code><a href="#@jaggr2/cdk-log-to-s3.ILogToS3Extension.property.keyPrefix">keyPrefix</a></code> | <code>string</code> | Normalised key prefix, e.g. 'logs/'. |
| <code><a href="#@jaggr2/cdk-log-to-s3.ILogToS3Extension.property.layer">layer</a></code> | <code>aws-cdk-lib.aws_lambda.ILayerVersion</code> | The layer carrying the extension binary. |
| <code><a href="#@jaggr2/cdk-log-to-s3.ILogToS3Extension.property.logsBucket">logsBucket</a></code> | <code>aws-cdk-lib.aws_s3.IBucket</code> | Bucket the extension writes to. |

---

##### `node`<sup>Required</sup> <a name="node" id="@jaggr2/cdk-log-to-s3.ILogToS3Extension.property.node"></a>

```typescript
public readonly node: Node;
```

- *Type:* constructs.Node

The tree node.

---

##### `architecture`<sup>Required</sup> <a name="architecture" id="@jaggr2/cdk-log-to-s3.ILogToS3Extension.property.architecture"></a>

```typescript
public readonly architecture: Architecture;
```

- *Type:* aws-cdk-lib.aws_lambda.Architecture

Architecture this layer was built for.

---

##### `environment`<sup>Required</sup> <a name="environment" id="@jaggr2/cdk-log-to-s3.ILogToS3Extension.property.environment"></a>

```typescript
public readonly environment: {[ key: string ]: string};
```

- *Type:* {[ key: string ]: string}

Environment variables attachTo() injects.

Exposed for manual wiring.

---

##### `keyPrefix`<sup>Required</sup> <a name="keyPrefix" id="@jaggr2/cdk-log-to-s3.ILogToS3Extension.property.keyPrefix"></a>

```typescript
public readonly keyPrefix: string;
```

- *Type:* string

Normalised key prefix, e.g. 'logs/'.

---

##### `layer`<sup>Required</sup> <a name="layer" id="@jaggr2/cdk-log-to-s3.ILogToS3Extension.property.layer"></a>

```typescript
public readonly layer: ILayerVersion;
```

- *Type:* aws-cdk-lib.aws_lambda.ILayerVersion

The layer carrying the extension binary.

---

##### `logsBucket`<sup>Required</sup> <a name="logsBucket" id="@jaggr2/cdk-log-to-s3.ILogToS3Extension.property.logsBucket"></a>

```typescript
public readonly logsBucket: IBucket;
```

- *Type:* aws-cdk-lib.aws_s3.IBucket

Bucket the extension writes to.

---

## Enums <a name="Enums" id="Enums"></a>

### LogCompression <a name="LogCompression" id="@jaggr2/cdk-log-to-s3.LogCompression"></a>

Compression codec used for the Parquet output.

#### Members <a name="Members" id="Members"></a>

| **Name** | **Description** |
| --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompression.SNAPPY">SNAPPY</a></code> | Good ratio at low CPU cost. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompression.ZSTD">ZSTD</a></code> | Smaller files, more CPU. |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompression.GZIP">GZIP</a></code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogCompression.UNCOMPRESSED">UNCOMPRESSED</a></code> | No compression. |

---

##### `SNAPPY` <a name="SNAPPY" id="@jaggr2/cdk-log-to-s3.LogCompression.SNAPPY"></a>

Good ratio at low CPU cost.

The default, and what the Glue table declares.

---


##### `ZSTD` <a name="ZSTD" id="@jaggr2/cdk-log-to-s3.LogCompression.ZSTD"></a>

Smaller files, more CPU.

---


##### `GZIP` <a name="GZIP" id="@jaggr2/cdk-log-to-s3.LogCompression.GZIP"></a>

---


##### `UNCOMPRESSED` <a name="UNCOMPRESSED" id="@jaggr2/cdk-log-to-s3.LogCompression.UNCOMPRESSED"></a>

No compression.

Mostly useful for debugging.

---


### LogLevel <a name="LogLevel" id="@jaggr2/cdk-log-to-s3.LogLevel"></a>

Minimum level a record must have to be written to S3.

#### Members <a name="Members" id="Members"></a>

| **Name** | **Description** |
| --- | --- |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogLevel.DEBUG">DEBUG</a></code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogLevel.INFO">INFO</a></code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogLevel.WARN">WARN</a></code> | *No description.* |
| <code><a href="#@jaggr2/cdk-log-to-s3.LogLevel.ERROR">ERROR</a></code> | *No description.* |

---

##### `DEBUG` <a name="DEBUG" id="@jaggr2/cdk-log-to-s3.LogLevel.DEBUG"></a>

---


##### `INFO` <a name="INFO" id="@jaggr2/cdk-log-to-s3.LogLevel.INFO"></a>

---


##### `WARN` <a name="WARN" id="@jaggr2/cdk-log-to-s3.LogLevel.WARN"></a>

---


##### `ERROR` <a name="ERROR" id="@jaggr2/cdk-log-to-s3.LogLevel.ERROR"></a>

---


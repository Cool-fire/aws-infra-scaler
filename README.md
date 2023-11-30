# AWS Infra Scaler

This is a CLI tool to scale up/down AWS infrastructure based on a YAML configuration file. This easy to use tool can be helpful while testing the AWS infrastructure scaling.


## Setup

1. Run ```make``` to build the binary. The binary will be created in the ```bin``` directory.
2. Set the AWS credentials in the environment variables. The tool will use these credentials to connect to AWS. Ensure that you have valid Region configured in the credentials.
3. Run the binary with ```--scale-up``` or ```--scale-down``` flag to scale up or down the infrastructure respectively. The tool will read the configuration from the ```config.yaml``` file in the current directory. You can also specify the path to the configuration file using the ```--config``` flag.

## Configuration

The CLI uses a YAML configuration file to read the configuration. Each configuration can have multiple scaling regions and each scaling region can have multiple services to scale. 

Each region and corresponding services scales independently. The configuration file should have the following structure:
```yaml
appName: "my-app" # Name of the application
assumedRoleArn: "arn:aws:iam::123456789012:role/my-app-role" # ARN of the IAM role to assume
scalingRegions: # List of regions to scale
  - region: "Region Name"
    serviceScaleConfigs: # List of services to scale in the region
        -service: "AWS Service Name"
        # Service specific configuration
```
check the [example config](./config.yaml) for more details.

### Service Specific Configuration

The service specific configuration depends on the service being scaled. The following services are supported:
1. DynamoDB
2. EC2
3. Elasticache
4. Kinesis

#### DynamoDB

DynamoDB's configuration considers Index and Table as separate entities. The configuration for DynamoDB is as follows:

```yaml
        service: "dynamodb"
        tableName: "ScaleUpTable"
        isIndex: false
        rcu:
          minProvisionedCapacity: 1
          maxProvisionedCapacity: 100
        wcu:
          minProvisionedCapacity: 1
          maxProvisionedCapacity: 100
```

The ```isIndex``` flag is used to specify whether the table is an index or not. If the table is an index, the ```tableName``` should be the name of the index and the ```isIndex``` flag should be set to ```true```.

#### Kinesis

```yaml
service: "kinesis"
streamArn: "arn:aws:kinesis:us-east-1:123456789012:stream/ScaleUpStream"
desiredShardCount: 1
```

Thing to note Kinesis can only be scaled to double the current shard count. So, if the current shard count is 1, the desired shard count can be 2 at max. If not followed, the scaling will fail.

### Elasticache

CLI supports scaling of Redis and Memcached. The configuration for Elasticache is as follows:

```yaml
service: "elasticache"
clusterId: "ScaleUpCluster"
engine: "redis" # redis or memcached
nodeCount: 1 
nodesToDelete: # Needed only for scaling down
    - "0001"
    - "0002"
```

The ```nodesToDelete``` is a required field and used only for scaling down, values defined for this fields won't be considered while scaling up, It is expected to keep empty list for scaling up. The node ids can be found in the AWS console.

### EC2

CLI supports scaling of EC2 ASG and EC2 instances. The configuration for EC2 is as follows:

```yaml
service: "ec2"
asgName: "ScaleUpASG"
minCount: 1
desiredCount: 1
maxCount: 1
```

### Error Handling

Each Region and corresponding services are scaled independently. Error in scaling one service or region will not affect the scaling of other services or regions.

The errors are accumulated per region for failed services and the CLI will print the error message along with identifierId and failed service name for each region at the end of the scaling process.

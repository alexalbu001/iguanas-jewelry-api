local env = std.extVar('env');
assert std.member(['staging', 'production', 'development'], env) : 'Invalid environment, must be one of ["staging", "production", "development"]';

local version = std.extVar('version');  // git sha string
assert std.length(version) == 40 : 'Received a version string that is shorter that 40 characters. Did you provide the full Git SHA?';

local aws_account_id = std.extVar('aws_account_id');
assert std.member(['264669033021'], aws_account_id) : 'Invalid AWS Account ID, must be one of ["264669033021"]';

local service = std.extVar('service');
assert std.member([
  'iguanas-jewelry-api',
], service) : 'Invalid service name, must be iguanas-jewelry-api';

local region = std.extVar('region');

{
  family: service + '-migrations',
  taskRoleArn: std.format('arn:aws:iam::%s:role/%s-task', [aws_account_id, service]),
  executionRoleArn: std.format('arn:aws:iam::%s:role/%s-executor', [aws_account_id, service]),
  networkMode: 'awsvpc',
  requiresCompatibilities: [
    'FARGATE',
  ],
  cpu: '256',
  memory: '512',
  runtimePlatform: {
    operatingSystemFamily: 'LINUX',
    cpuArchitecture: 'X86_64',
  },
  containerDefinitions: [
    {
      essential: true,
      name: service + '-migrations',
      image: std.format('%s.dkr.ecr.%s.amazonaws.com/iguanas-jewelry-api:%s', [aws_account_id, region, version]),
      entryPoint: ['/bin/sh', '-c'],
      command: ['migrate -database "$DATABASE_URL" -path migrations up && echo "Migrations completed successfully"'],
      environment: [
        {
          name: 'ENV',
          value: env,
        }
      ],
      linuxParameters: {
        initProcessEnabled: true,
      },
      secrets: [
        {
          name: 'DATABASE_URL',
          valueFrom: 'DATABASE_URL',
        }
      ],
      readonlyRootFilesystem: false,
      interactive: false,
      pseudoTerminal: false,
      logConfiguration: {
        logDriver: 'awslogs',
        options: {
          'awslogs-group': std.format('/ecs/%s-migrations', service),
          'awslogs-region': region,
          'awslogs-stream-prefix': 'ecs',
          'awslogs-create-group': 'true',
        },
      },
    }
  ],
  tags: [
    {
      key: 'owner',
      value: 'alexalbu001@gmail.com',
    },
    {
      key: 'product',
      value: 'jewelry-ecommerce',
    },
    {
      key: 'env',
      value: env,
    },
    {
      key: 'service',
      value: 'iguanas-jewelry-api-migrations',
    },
  ],
}


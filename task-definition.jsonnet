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
  family: service,
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
      name: service,
      image: std.format('%s.dkr.ecr.%s.amazonaws.com/iguanas-jewelry-api:%s', [aws_account_id, region, version]),
      portMappings: [
        {
          containerPort: 8080,
          protocol: 'tcp',
          name: 'api'
        }
      ],
      entryPoint: ['./main'],
      command: [],
      healthCheck: {
        command: [
          'CMD-SHELL',
          'curl -f http://localhost:8080/health || exit 1'
        ],
        interval: 30,
        timeout: 5,
        retries: 3,
        startPeriod: 60
      },
      environment: [
        {
          name: 'ENV',
          value: env,
        },
        {
          name: 'VERSION',
          value: version,
        },
        {
          name: 'PORT',
          value: '8080',
        },
        {
          name: 'LOG_LEVEL',
          value: 'info',
        },
        {
          name: 'LOG_FORMAT',
          value: 'json',
        },
        {
          name: 'WORKER_MODE',
          value: 'scheduler',
        },
        {
          name: 'IMAGE_STORAGE_MODE',
          value: 's3',
        },
        {
          name: 'CORS_ALLOWED_ORIGINS',
          value: 'https://iguanasjewellery.com,https://www.iguanasjewellery.com,https://admin.iguanasjewellery.com',
        },
        {
          name: 'FROM_EMAIL',
          value: 'alexalbu001@gmail.com',
        },
        {
          name: 'FROM_NAME',
          value: 'Iguanas Jewelry',
        },
        {
          name: 'IMAGE_STORAGE_BUCKET',
          value:"iguanas-jewelry-assets"
        },
        {
          name: 'IMAGE_STORAGE_BASE_URL',
          value: 'https://www.iguanasjewellery.com'
        },
        {
          name: 'IMAGE_STORAGE_REGION',
          value:"eu-west-1"
        },
        {
          name: 'ADMIN_EMAIL',
          value:"alexalbu001@gmail.com"
        },
        {
          name: 'ADMIN_ORIGIN',
          value:"https://admin.iguanasjewellery.com"
        },
        {
          name: 'REDIRECT_URL',
          value:"https://api.iguanasjewellery.com/auth/google/callback"
        },
      ],
      linuxParameters: {
        initProcessEnabled: true,  // Used for ecs execute-command to avoid SSM agent child processes becoming orphaned.
      },
      secrets: [
        {
          name: 'JWT_SECRET',
          valueFrom: std.format('/secrets/iguanas-jewelry/JWT_SECRET', []),
        },
        {
          name: 'DATABASE_URL',
          valueFrom: 'DATABASE_URL',
        },
        {
          name: 'REDIS_URL',
          valueFrom: 'REDIS_URL',
        },
        {
          name: 'STRIPE_SECRET_KEY',
          valueFrom: 'STRIPE_SECRET_KEY',
        },
        {
          name: 'STRIPE_WEBHOOK_SECRET',
          valueFrom: 'STRIPE_WEBHOOK_SECRET',
        },
        {
          name: 'GOOGLE_CLIENT_ID',
          valueFrom: 'GOOGLE_CLIENT_ID',
        },
        {
          name: 'GOOGLE_CLIENT_SECRET',
          valueFrom: 'GOOGLE_CLIENT_SECRET',
        },
        {
          name: 'SENDGRID_API_KEY',
          valueFrom: 'SENDGRID_API_KEY',
        }
      ],
      readonlyRootFilesystem: false,
      interactive: false,
      pseudoTerminal: false,
      ulimits: [
        {
          name: 'nofile',
          softLimit: 65536,
          hardLimit: 65536,
        },
      ],
      logConfiguration: {
        logDriver: 'awslogs',
        options: {
          'awslogs-group': std.format('/ecs/%s', service),
          'awslogs-region': region,
          'awslogs-stream-prefix': 'ecs',
          'awslogs-create-group': 'true',
        },
      },
    },
    {
      name: 'healthcheck',  // Service Connect Health Checks Workaround https://github.com/aws/containers-roadmap/issues/2334
      image: 'public.ecr.aws/docker/library/alpine:edge',
      essential: false,
      dependsOn: [
        {
          containerName: service,
          condition: 'HEALTHY',
        },
      ],
    },
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
      value: 'iguanas-jewelry-api',
    },
  ],
}

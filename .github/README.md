# GitHub Actions Workflows

This directory contains GitHub Actions workflows for automated testing, building, and documentation deployment.

## Workflows

### 1. Tests (`test.yml`)
Runs comprehensive tests including:
- Unit tests
- Integration tests  
- End-to-end tests

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

**Features:**
- PostgreSQL service container for database tests
- Test result artifacts upload
- Coverage reporting

### 2. Documentation Deployment (`docs.yml`)
Automatically builds and deploys documentation to GitHub Pages.

**Triggers:**
- Push to `main` branch
- Manual workflow dispatch

**Features:**
- Generates documentation using `make docs`
- Deploys to GitHub Pages
- Includes interactive HTML documentation

### 3. Continuous Integration (`ci.yml`)
Comprehensive CI pipeline including:
- Code linting with golangci-lint
- Multi-platform builds (linux/amd64, linux/arm64)
- Security vulnerability scanning
- Test coverage reporting

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

### 4. Docker Build Test (`docker.yml`)
Tests Docker image building and Docker Compose functionality.

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

**Features:**
- Builds server and playground Docker images
- Tests Docker Compose configuration
- Uses Docker Buildx for efficient builds
- GitHub Actions cache for faster builds

## Configuration

### Linting Configuration
The `.golangci.yml` file configures code quality checks including:
- Code formatting (gofmt, goimports)
- Static analysis (govet, staticcheck)
- Security scanning (gosec)
- Code complexity analysis (gocyclo)
- Duplicate code detection (dupl)

### Test Environment
Tests use PostgreSQL 15 with the following configuration:
- Database: `test_db`
- User: `postgres`
- Password: `postgres`
- Port: `5432`

## Usage

### Local Development
```bash
# Run tests locally
make test

# Run linting
golangci-lint run

# Generate documentation
make docs

# Test Docker builds
docker compose -f docker-compose.test.yml build
```

### GitHub Pages
Documentation is automatically deployed to GitHub Pages when:
1. Code is pushed to the `main` branch
2. The `docs.yml` workflow completes successfully

The documentation will be available at: `https://[username].github.io/[repository-name]/`

### Manual Workflow Triggers
You can manually trigger workflows from the GitHub Actions tab:
1. Go to your repository on GitHub
2. Click on the "Actions" tab
3. Select the workflow you want to run
4. Click "Run workflow"

## Artifacts

The workflows generate several artifacts:
- **Test Results**: Available for 30 days
- **Coverage Reports**: HTML and raw coverage data
- **Build Binaries**: For multiple platforms
- **Documentation**: Deployed to GitHub Pages

## Troubleshooting

### Common Issues

1. **Tests failing due to database connection**
   - Ensure PostgreSQL service is running
   - Check database connection string in environment variables

2. **Linting errors**
   - Run `golangci-lint run` locally to see issues
   - Update `.golangci.yml` to exclude specific rules if needed

3. **Documentation not deploying**
   - Check that GitHub Pages is enabled in repository settings
   - Verify the `docs` directory contains generated files
   - Check workflow logs for generation errors

4. **Docker builds failing**
   - Check Dockerfile syntax
   - Verify Docker Compose configuration
   - Check for missing dependencies

### Environment Variables
The following environment variables are used:
- `DATABASE_URL`: PostgreSQL connection string for tests
- `GITHUB_TOKEN`: Automatically provided by GitHub Actions

## Contributing

When adding new workflows or modifying existing ones:
1. Test locally first using `act` or similar tools
2. Ensure all required permissions are set
3. Update this README with any changes
4. Test the workflow on a feature branch before merging 
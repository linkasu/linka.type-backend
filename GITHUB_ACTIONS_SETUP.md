# GitHub Actions Setup

This document describes the GitHub Actions workflows that have been added to the Linka Type Backend project.

## 📋 Overview

The following GitHub Actions workflows have been configured:

1. **Tests** (`test.yml`) - Comprehensive testing pipeline
2. **Documentation Deployment** (`docs.yml`) - GitHub Pages documentation
3. **Continuous Integration** (`ci.yml`) - Code quality and security checks
4. **Docker Build Test** (`docker.yml`) - Docker image building and testing

## 🚀 Workflows Details

### 1. Tests Workflow (`test.yml`)

**Purpose:** Run all tests including unit, integration, and e2e tests

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

**Features:**
- PostgreSQL 15 service container for database tests
- Test result artifacts upload (30 days retention)
- Coverage reporting
- Uses existing Makefile commands

**Environment:**
- Go 1.24.5
- PostgreSQL 15 with health checks
- Database: `test_db`, User: `postgres`, Password: `postgres`

### 2. Documentation Deployment (`docs.yml`)

**Purpose:** Automatically deploy documentation to GitHub Pages

**Triggers:**
- Push to `main` branch
- Manual workflow dispatch

**Features:**
- Generates documentation using `make docs`
- Deploys to GitHub Pages
- Includes interactive HTML documentation
- Beautiful landing page with navigation

**Output:** Documentation available at `https://[username].github.io/[repository-name]/`

### 3. Continuous Integration (`ci.yml`)

**Purpose:** Comprehensive code quality and security checks

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

**Features:**
- Code linting with golangci-lint
- Multi-platform builds (linux/amd64, linux/arm64)
- Security vulnerability scanning
- Test coverage reporting
- Build artifacts upload

### 4. Docker Build Test (`docker.yml`)

**Purpose:** Test Docker image building and Docker Compose functionality

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

**Features:**
- Builds both server and playground images
- Tests Docker Compose configuration
- Uses Docker Buildx for efficient builds
- GitHub Actions cache for faster builds
- No external registry publishing

## ⚙️ Configuration Files

### `.golangci.yml`
Code quality configuration including:
- Code formatting (gofmt, goimports)
- Static analysis (govet, staticcheck)
- Security scanning (gosec)
- Code complexity analysis (gocyclo)
- Duplicate code detection (dupl)

### `docs/index.html`
Beautiful landing page for GitHub Pages with:
- Modern, responsive design
- Navigation to all documentation files
- Project overview and features
- Links to API docs, import guides, and generated specs

## 🔧 Setup Requirements

### GitHub Repository Settings

1. **Enable GitHub Pages:**
   - Go to Settings → Pages
   - Source: "GitHub Actions"
   - Branch: `main`

### Local Development

The workflows use existing Makefile commands:
```bash
# Run tests
make test

# Generate documentation
make docs

# Run linting
golangci-lint run

# Test Docker builds
docker compose -f docker-compose.test.yml build
```

## 📊 Workflow Status

You can monitor workflow status at:
- GitHub repository → Actions tab
- Each workflow shows detailed logs and results
- Artifacts are available for download

## 🛠️ Customization

### Adding New Tests
1. Add test files to appropriate directories (`tests/unit/`, `tests/integration/`, `tests/e2e/`)
2. Update Makefile if needed
3. Tests will automatically run in CI

### Modifying Documentation
1. Update documentation generation in `cmd/docs/`
2. Modify `docs/index.html` for landing page changes
3. Documentation will auto-deploy on main branch pushes

### Adding New Workflows
1. Create new `.yml` file in `.github/workflows/`
2. Follow existing patterns for consistency
3. Update `.github/README.md` with documentation

## 🔍 Troubleshooting

### Common Issues

1. **Tests failing:**
   - Check PostgreSQL service is running
   - Verify database connection string
   - Check test logs for specific errors

2. **Documentation not deploying:**
   - Ensure GitHub Pages is enabled
   - Check `docs` directory contains files
   - Verify workflow permissions

3. **Docker builds failing:**
   - Check Dockerfile syntax
   - Verify Docker Compose configuration
   - Check for missing dependencies

### Environment Variables

The workflows use these environment variables:
- `DATABASE_URL`: PostgreSQL connection for tests
- `GITHUB_TOKEN`: Automatically provided

## 📈 Benefits

This setup provides:

1. **Automated Testing:** All tests run on every push/PR
2. **Code Quality:** Linting and security scanning
3. **Documentation:** Auto-deployed to GitHub Pages
4. **Docker Testing:** Automated Docker image building and testing
5. **Multi-platform:** Support for different architectures
6. **Artifacts:** Test results and coverage reports
7. **Security:** Vulnerability scanning and secure practices

## 🎯 Next Steps

1. Push these changes to your repository
2. Enable GitHub Pages in repository settings
3. Monitor the first workflow runs
4. Customize workflows as needed for your specific requirements

The workflows are designed to work with your existing project structure and Makefile, providing a comprehensive CI/CD pipeline for the Linka Type Backend project. 
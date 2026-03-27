# Contributing Guidelines

Thank you for your interest in contributing to OpenGIN-Org! We welcome contributions from everyone. This document provides guidelines and best practices for contributing.

This project is a command-line tool for managing and processing organisational structure data, specifically designed for tracking government ministries, departments, and personnel appointments in Sri Lanka.

## Code of Conduct

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## How to Contribute

There are many ways to contribute to this project:

- **Report Bugs**: Submit bug reports with detailed information
- **Suggest Features**: Propose new features or improvements
- **Improve Documentation**: Fix typos, clarify explanations, add examples
- **Submit Code**: Fix bugs or implement new features
- **Review Pull Requests**: Help review and test contributions from others

## Getting Started

### Prerequisites

- Go 1.24 or above
- Access to the required API endpoints
- Git

### Development Setup

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR-USERNAME/opengin-org.git
   cd opengin-org
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Build the project:
   ```bash
   go build -o orgchart cmd/main.go
   ```

For detailed development instructions, API setups, and architecture details, please refer to [DEVELOPMENT.md](DEVELOPMENT.md).

## Making Changes

### Branching Strategy

- `feature/` - for new features
- `fix/` - for bug fixes
- `docs/` - for documentation changes

Create a topic branch from the main branch for your changes:

```bash
git checkout -b feature/your-feature-name
```

### Commit Messages

Write clear and meaningful commit messages. We recommend following this format:

```
[TYPE] Short description (max 50 chars)

Longer description if needed. Explain the "why" behind the change,
not just the "what". Reference any related issues.

Fixes #123
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

### Coding Standards

- Follow standard Go conventions (e.g., `gofmt`).
- Ensure code is well-documented.

### Testing

All changes should include appropriate tests. Run the test suite before submitting:

```bash
go test ./tests
```

## Submitting Changes

### Pull Request Process

1. Ensure your code follows the project's coding standards
2. Update documentation if needed
3. Add or update tests as appropriate
4. Run the full test suite and ensure it passes
5. Push your branch and create a Pull Request

### Pull Request Guidelines

- Provide a clear title and description
- Reference any related issues (e.g., "Fixes #123")
- Keep changes focused and atomic
- Be responsive to feedback and review comments

### Review Process

- PRs require at least one approval from a maintainer
- CI checks must pass
- Changes may be requested before merging

## Communication

- GitHub Issues: For bug reports and feature requests
- Email: [contact@datafoundation.lk](mailto:contact@datafoundation.lk)

## Recognition

We value all contributions and appreciate your effort to improve this project!

## Additional Resources

- [Project Documentation](README.md)
- [Development Guide](DEVELOPMENT.md)

---

*These guidelines are inspired by the [Apache Way](https://www.apache.org/theapacheway/) and [Open Source Guides](https://opensource.guide/).*

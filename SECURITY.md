# Security Policy


## Reporting a Vulnerability

We take the security of FM-Dash seriously. If you discover a security vulnerability, please follow these steps:

### 🔒 Private Disclosure

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them privately here:
**GitHub Security Advisories**
   - Go to the [Security tab](https://github.com/LiamHardman/fm-dash/security/advisories/new)
   - Click "Report a vulnerability"
   - Fill out the form with details


### 📋 What to Include

When reporting a vulnerability, please include as much of the following information as possible:

- **Description**: A clear description of the vulnerability
- **Impact**: The potential impact and severity
- **Reproduction**: Step-by-step instructions to reproduce the issue
- **Affected Components**: Which parts of the application are affected
- **Proposed Fix**: If you have suggestions for how to fix the issue
- **Your Environment**: Version, deployment method, browser, etc.

### 🕐 Response Timeline

We aim to respond to security reports according to the following timeline:

- **Initial Response**: Within 24-48 hours
- **Triage**: Within 72 hours
- **Fix Development**: 1-7 days (depending on severity)
- **Release**: As soon as possible after fix is ready

### 📊 Severity Levels

We classify vulnerabilities using the following severity levels:

| Severity | Description | Response Time |
|----------|-------------|---------------|
| **Critical** | Remote code execution, data breach, authentication bypass | 24 hours |
| **High** | Privilege escalation, significant data exposure | 48 hours |
| **Medium** | Limited data exposure, denial of service | 72 hours |
| **Low** | Information disclosure, minor security issues | 1 week |

## Security Best Practices

### For Users

- **Keep Updated**: Always use the latest supported version
- **Secure Deployment**: Follow deployment security guidelines
- **Environment Variables**: Never commit secrets to version control
- **Access Control**: Implement proper access controls in production
- **HTTPS**: Always use HTTPS in production environments

### For Developers

- **Code Review**: All code changes must be reviewed
- **Dependency Updates**: Keep dependencies updated and scan for vulnerabilities
- **Input Validation**: Validate and sanitize all user inputs
- **Authentication**: Implement secure authentication mechanisms
- **Logging**: Avoid logging sensitive information

## Security Features

FM-Dash includes several built-in security features:

- **Input Validation**: Server-side validation of all inputs
- **CORS Protection**: Configurable CORS policies
- **Rate Limiting**: Built-in rate limiting capabilities
- **Security Headers**: Automatic security headers in responses
- **Secure Defaults**: Secure configuration out of the box

## Known Security Considerations

- **File Uploads**: HTML file uploads are processed server-side - ensure files are from trusted sources
- **S3 Storage**: When using S3 storage, ensure proper IAM policies and bucket permissions
- **Environment Variables**: Several environment variables contain sensitive information (S3 credentials, etc.)

## Security Updates

Security updates will be:

- Released as patch versions (e.g., 1.4.1)
- Documented in the changelog with security impact
- Announced through GitHub releases
- Tagged with security labels in commit messages


## Contact

For security-related questions or concerns:
- GitHub: [@LiamHardman](https://github.com/LiamHardman)
- Security Advisories: [GitHub Security Tab](https://github.com/LiamHardman/fm-dash/security)

---
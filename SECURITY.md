# Security Policy

## Supported Versions

We provide security updates for the following versions of Citadel Agent:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | ✅ Latest          |
| < 1.0   | ❌ Unsupported     |

## Reporting a Vulnerability

We take the security of Citadel Agent seriously. If you believe you have found a security vulnerability, please follow these steps:

### Disclosure Process

1. **Do not report security vulnerabilities through public GitHub issues**
2. Email your findings to [security@citadel-agent.com](mailto:security@citadel-agent.com)
3. Include a detailed description of the vulnerability
4. Include steps to reproduce the issue (if possible)
5. Include potential impact of the vulnerability

### What to Expect

- You will receive an acknowledgment of your report within 48 hours
- We will investigate and respond with our findings and plans within 7 days
- If we decide to fix the issue, we will notify you when the fix is released
- If the issue is declined, we will explain our reasoning

### Responsible Disclosure Guidelines

We ask that you:
- Give us reasonable time to investigate and address the issue before making it public
- Do not use the vulnerability to access unauthorized systems or data
- Do not use attacks on physical security, social engineering, distributed denial of service, spam, or applications of third parties
- Do not interact with real user accounts or data in a production environment
- Demonstrate good faith by not attempting to access or modify other users' data

## Security Best Practices

When using Citadel Agent, please follow these security best practices:

### For Administrators
- Keep Citadel Agent updated to the latest version
- Use strong, unique passwords for all accounts
- Enable and enforce multi-factor authentication where possible
- Regularly backup your data
- Monitor logs for unusual activity
- Use HTTPS in production environments
- Regularly rotate API keys and secrets
- Implement appropriate network security measures

### For Developers
- Validate all input data
- Use parameterized queries to prevent SQL injection
- Implement proper authentication and authorization checks
- Follow the principle of least privilege
- Regularly update dependencies
- Use secure coding practices

## Security Features

Citadel Agent includes several built-in security features:

- **Authentication**: JWT-based authentication with configurable expiration
- **Authorization**: Role-based access control (RBAC)
- **Sandboxing**: Isolated execution environments for workflow nodes
- **Encryption**: TLS for network traffic, option for data at rest
- **Rate Limiting**: Protection against abuse and DoS attacks
- **Input Validation**: Sanitization of user inputs

## Compliance

Citadel Agent is designed to help you meet common compliance requirements:
- GDPR - Data privacy and user rights features
- SOC 2 - Secure data handling and access controls
- ISO 27001 - Security management framework compatibility

## Questions?

If you have questions about the security of Citadel Agent, please contact us at [security@citadel-agent.com](mailto:security@citadel-agent.com).
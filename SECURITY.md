# Security Policy

## Supported Versions

We take security seriously and provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| 0.5.x   | :x:                |
| < 0.5   | :x:                |

Starting with v1.0.0, we follow semantic versioning and provide security updates for all v1.x releases.

## Reporting a Vulnerability

If you discover a security vulnerability in go-cyacd, please report it by:

**Opening a GitHub Issue:** https://github.com/moffa90/go-cyacd/issues/new

Use the "Bug Report" template and add the **security** label to your issue.

### What to Include

Please include the following information in your report:

- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact and severity
- Suggested fix (if you have one)

### Response Timeline

- **Initial Response:** Within 48 hours
- **Status Update:** Within 7 days
- **Fix Timeline:** Depends on severity
  - Critical: Within 7 days
  - High: Within 14 days
  - Medium: Within 30 days
  - Low: Next minor release

### Disclosure Policy

- Please do not publicly disclose the vulnerability until we've had a chance to address it
- We will credit you in the security advisory (unless you prefer to remain anonymous)
- We will coordinate the disclosure timeline with you

### Security Advisories

Security advisories will be published at:
- GitHub Security Advisories: https://github.com/moffa90/go-cyacd/security/advisories
- CHANGELOG.md with detailed information
- Release notes for patched versions

## Security Considerations for Users

When using go-cyacd in your applications:

### Bootloader Keys

- **Never hardcode bootloader keys** in your source code
- Store keys securely using environment variables or secure key management systems
- Use different keys for development and production devices

### Firmware Files

- **Validate firmware files** before programming
- Verify firmware sources and checksums
- Implement access controls for firmware storage

### Device Communication

- **Implement timeouts** using context.Context to prevent hanging operations
- **Validate device responses** before processing
- **Handle errors appropriately** to prevent undefined states

### Network/Remote Programming

If using go-cyacd over a network:

- **Use encrypted channels** (TLS/SSL) for firmware transmission
- **Authenticate devices** before programming
- **Implement rate limiting** to prevent abuse
- **Log all programming operations** for audit trails

## Known Security Limitations

### Hardware Security

This library provides **software-level protocol implementation only**. It does not:

- Implement cryptographic firmware signing (device-dependent)
- Provide hardware-level security features
- Prevent physical attacks on devices
- Manage device-specific security keys

Refer to Infineon's PSoC security documentation for hardware-level security features.

### Input Validation

The library validates:
- ✅ Firmware file format (.cyacd parsing)
- ✅ Packet checksums
- ✅ Device silicon ID matching
- ✅ Row boundaries and sizes

The library does NOT validate:
- ❌ Firmware content or behavior
- ❌ Malicious firmware payloads
- ❌ Firmware authenticity or signatures

**Recommendation:** Implement firmware verification and signing at the application level.

## Security Best Practices

### Production Deployments

```go
// ✅ GOOD: Secure key management
key := getKeyFromSecureVault() // Your secure key storage

// ✅ GOOD: Timeout protection
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

// ✅ GOOD: Error handling
err := prog.Program(ctx, fw, key)
if err != nil {
    log.Printf("Programming failed: %v", err)
    // Handle error appropriately
}
```

```go
// ❌ BAD: Hardcoded keys
key := []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F} // Don't do this!

// ❌ BAD: No timeout
err := prog.Program(context.Background(), fw, key) // Could hang forever

// ❌ BAD: Ignoring errors
_ = prog.Program(ctx, fw, key) // Device may be in undefined state
```

## Third-Party Dependencies

go-cyacd has **zero external dependencies** beyond the Go standard library, minimizing the attack surface and dependency chain vulnerabilities.

## Security Updates

Subscribe to security updates:
- Watch this repository for security advisories
- Follow releases: https://github.com/moffa90/go-cyacd/releases
- Check CHANGELOG.md for security fixes

## Questions

For security-related questions that are not vulnerabilities, please open a GitHub Discussion or submit an issue.

---

**Last Updated:** 2025-11-18 (v1.0.0)

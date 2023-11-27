# Security

## Human access

### Unix Socket

Several techniques are used to ensure the security model.

1. Unix ACL. By default, socket access is limited to the user being started. (the user can change this)
2. Inside the application, it is checked who is making the request by UID. If it is not the same as the user being launched and not root (uid 0) - the request will be rejected. (cannot be disabled or skipped)

## Robot Access

It is not possible to use Yubikey for robots to access equipment. For this purpose, short-lived certificates are used. (for example 1 day)

To do this, a separate client is begin developed, called roboclient.

The certificate access policy is inherited from the regular client, with a few exceptions.

1. The certificate and private key is always stored in memory. Each launch generates a new certificate and key, i.e. it will never be saved to the file system.
2. The client can request that the certificate be revoked. (for example, when restarting)

Do you need access for different applications with different users in the OS? Run different copies of roboclient for each user! (no shared certificates!)

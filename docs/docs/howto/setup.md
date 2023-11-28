# Setup

## Certificate Authority Infrastucture

``` mermaid
flowchart TD
    A[Company Root CA] --> B[OneAuth Root CA]
    B --> C[OneAuth Intermediate CA]
    C --> D(Signed Certificate)
```

### Company Root CA

The Company Root CA is the root of trust for all certificates issued by the company. It is used to sign the OneAuth Root CA.

### OneAuth Root CA

The OneAuth Root CA is the root of trust for all certificates issued by OneAuth. It is used to sign the OneAuth Intermediate CA. This certificate is used for checking the validity in the servers.

You need to generate it and store it in a safe and secure place.

### OneAuth Intermediate CA

The OneAuth Intermediate CA is the intermediate CA used to sign the end-entry Certificate. This Certificate need to add to the OneAuth Server.

In case of compromentation, you can replace it, since the trust check occurs at a higher level.

!!! info
    Consumers are required to learn to rotate. Be able to periodically go for fresh public keys, this will allow you to react to the situation less painfully in the event of a hack, since the rotation mechanism will be worked out.

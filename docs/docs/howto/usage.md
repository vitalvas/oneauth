# Usage

## SSH agent

List all available keys:

```bash
SSH_AUTH_SOCK=~/.oneauth/ssh-agent.sock ssh-add -L
```

### SSH client configuration

```bash
Host *.srv.example.com
    # for authentication with ssh-agent to hosts
    IdentityAgent ~/.oneauth/ssh-agent.sock

Host *.bastion.example.com
    # for authentication with ssh-agent to bastion host
    IdentityAgent ~/.oneauth/ssh-agent.sock
    # for authentication with ssh-agent from bastion host to hosts (forwarding agent)
    ForwardAgent ~/.oneauth/ssh-agent.sock
```

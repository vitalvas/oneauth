# Motivation

Over a long period of administering various equipment, many desires have accumulated so that access can be easily controlled. The most suitable solution is to use certificates from your certificate authority.

SSH certificates are more convenient than SSH keys.
But it is impossible to use certificates everywhere (for example, on network equipment), so it is necessary to have support for both the certificates themselves and traditional keys.

## General requirements

* Give time-limited access
* Give access to all servers
* Give access to a group of servers
* Possibility of using role-based or user-based access control
* Stop deploying personal keys manually or using automation systems
* Stop using passwords
* Protect yourself from the possibility of key theft
* ...and much more...

## Design philosophy

Most often, switching to authentication using SSH certificates involves replacing the SSH client with a custom one.
In our case, this was not an option at all and I chose the path of implementing an SSH agent because it was important to me:

* Don't break anything (for example: some applications create a client on their own, and it is not possible to replace it with your own). Anything that can communicate with SSH agents should continue to work as before.
* Don't change the user experience, workflow and habits.
* Don't radically change the server side configuration.

# gerberos

gerberos scans sources for lines matching regular expressions and containing IPv4 or IPv6 addresses and performs actions on those addresses.
Possible sources are (not necessarily existant) non-directory files, systemd journals, kernel messages, and standard outputs of arbitrary processes.
Addresses can be logged or added to ipsets or nft rulesets that gerberos will manage autonomously.

Minimal additional logic is applied. This is to adhere to the [Unix philosophy](https://en.wikipedia.org/wiki/Unix_philosophy), but impacts gerberos' out-of-the-box usefulness for specific use cases when compared to tools like [fail2ban](https://github.com/fail2ban/fail2ban).

## Current status

While this repository does not get much commits nowerdays and might go without major changes even over longer periods spanning multiple years, that is because the software is pretty much feature complete (and *possibly* bugfree) and doing its work quietly in the background. In fact, it is run on multiple servers for years now without any problems. That said, the software is still maintained and this text will stay here as long as the software is maintained.

If you find any bugs or want any features (without breaking the philosophy of minimal additional logic), do not hesitate to fill a new issue.

## Requirements

### ipset backend

- ipset 6.34
- iptables 1.6.1

### nft backend

- nftables v0.9.3 (tested on Ubuntu 20.04)

### Development only

- Go 1.24
- GNU Make 4.3 (optional)
- pgrep (system tests only, optional)

## Build

`make build`

## Test

### Unit tests only

`make test`

### Unit and system tests

Requires ipset, iptables, and nftables to be installed.

`make test_system`

## Example configuration file (TOML)

See [gerberos.toml](gerberos.toml).

## Example systemd service file

See [gerberos.service](gerberos.service).

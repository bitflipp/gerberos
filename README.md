# gerberos

Simple (log) file watcher

## Example configuration file (TOML)

```toml
[rules]
    [rules.apache-fuzzing-4]
    # Available sources are
    # - ["file", "<path to non-directory file>"]
    # - ["systemd", "<name of systemd service>"]
    source = ["file", "/var/log/apache2/access.log"]
    # Must contain exactly one of "%ip4%" and "%ip6%".
    # It will be replaced with a subexpression named
    # "host" matching IPv4 and IPv6 addresses, respec-
    # tively.
    regexp = "%ip4%.*40(0|8) 0 \"-\" \"-\""
    # Available actions are
    # - ["ban", "<value parsable by time.ParseDuration>"]
    # - ["log"]
    action = ["ban", "1h"]

    [rules.apache-fuzzing-6]
    source = ["file", "/var/log/apache2/access.log"]
    regexp = "%ip6%.*40(0|8) 0 \"-\" \"-\""
    action = ["ban", "1h"]
```

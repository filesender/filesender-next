# systemd

To install the `filesender` binary in `/usr/local/bin/filesender`, run this
from the "root" project folder:

```bash
$ sudo make install
```

Copy/setup systemd:

```bash
$ sudo cp systemd/filesender.service /etc/systemd/system
$ sudo systemctl daemon-reload
```

Start `filesender`:

```bash
$ sudo systemctl start filesender
```

Check the log:

```bash
$ journalctl -f -t filesender
```

The output looks like this:

```bash
$ journalctl -f -t filesender
Mär 06 19:55:10 fralen filesender[176362]: 2025/03/06 19:55:10 Using database: /var/lib/filesender/filesender.db
Mär 06 19:55:10 fralen filesender[176362]: 2025/03/06 19:55:10 Database connected and migrated successfully.
Mär 06 19:55:10 fralen filesender[176362]: 2025/03/06 19:55:10 HTTP server listening on 127.0.0.1:8080
```

If you want to change the listening port:

```bash
$ sudo systemctl edit filesender
```

Add the following to the top of the file as indicated:

```ini
[Service]
Environment=LISTEN=[::]:9000
```

Then restart `filesender`:

```bash
$ sudo systemctl restart filesender
```

This will allow all connections coming from anywhere on IPv4 and IPv6 on port
9000.

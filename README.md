![H-Bank](assets/banner.png)
<h1 align="center">H-Bank API</h1>

![License](https://img.shields.io/github/license/Bananenpro/hbank-api)
![Go Version](https://img.shields.io/github/go-mod/go-version/Bananenpro/hbank-api)
![Issues](https://img.shields.io/github/issues/Bananenpro/hbank-api)
![Stars](https://img.shields.io/github/stars/Bananenpro/hbank-api)

## Description

The API for the [H-Bank](https://github.com/juho05/hbank-web) application.

## Setup

### Cloning the repo

```bash
git clone https://github.com/juho05/hbank-api.git
cd hbank-api
```

### Installing dependencies

- [Go](https://go.dev) 1.17+

#### Windows

Using [chocolatey](https://chocolatey.org/) as administrator:

```powershell
choco install golang
```

#### Arch Linux

```bash
sudo pacman -S go
```

### Running

Make sure you are in the root directory of the project.

#### API

The api and webserver application.

```bash
make run_api
```

#### Payment plans

The service that executes payment plans.

```bash
make run_payment_plans
```

### Building

#### API

```bash
make build_api
```

The resulting binary can be found at `bin/hbank-api`.

#### Payment plans

```bash
make build_payment_plans
```

The resulting binary can be found at `bin/hbank-payment-plans`.

### Runt tests

```bash
make test
```

### Clean

```bash
make clean
```

### Deploying

#### Linux distribution with systemd

First [install all dependencies](#installing-dependencies) and [build](#building) the `hbank-api` and `hbank-payment-plans` binaries.

Create a new `hbank` directory in the home directory of the current user:
```bash
cd
mkdir hbank
```

Copy the binaries into the created directory:
```bash
cp <repo-dir>/bin/hbank-api ~/hbank
cp <repo-dir>/bin/hbank-payment-plans ~/hbank
```

Copy other required directories and files into the `hbank` directory:
```bash
cp -r <repo-dir>/assets ~/hbank
cp -r <repo-dir>/templates ~/hbank
cp -r <repo-dir>/translations ~/hbank
```

Create a configuration file named `config.json` in the same directory.

Example configuration:
```jsonc
{
  "domainName": "hbank.example",
  "baseURL": "https://hbank.example",
  "frontendURL": "https://hbank.example",
  "emailEnabled": true,
  "emailHost": "smtp.gmail.com",
  "emailPort": 587,
  "emailUsername": "example@gmail.com",
  "emailPassword": "verysecurepassword",
  "idProvider": "https://id.example.com",
  "clientID": "asöldkfjasödkfljasdöfkljasdf",
  "clientSecret": "asdfjkasdöfkljasdföklajsdföaslkdjf"
}
```

Create a systemd service file for `hbank-api`:
```bash
sudoedit /etc/systemd/system/hbank-api.service
```
with the following content (make sure to replace `<user>` with your username):
```
[Unit]
Description=H-Bank API
After=network.target

[Service]
ExecStart=/home/<user>/hbank/hbank-api
WorkingDirectory=/home/<user>/hbank
StandardOutput=inherit
StandardError=inherit
Restart=always
User=<user>

[Install]
WantedBy=multi-user.target
```

Create a systemd service file for `hbank-payment-plans`:
```bash
sudoedit /etc/systemd/system/hbank-payment-plans.service
```
with the following content:
```
[Unit]
Description=H-Bank Payment Plans
After=network.target

[Service]
ExecStart=/home/<user>/hbank/hbank-payment-plans
WorkingDirectory=/home/<user>/hbank
StandardOutput=inherit
StandardError=inherit
User=<user>
Type=oneshot
```

Create a systemd timer file for `hbank-payment-plans`:
```bash
sudoedit /etc/systemd/system/hbank-payment-plans.timer
```
with the following content:
```
[Unit]
Description=Run H-Bank Payment Plans Daily

[Timer]
OnCalendar=daily
Persistent=true
RandomizedDelaySec=1h

[Install]
WantedBy=timers.target
```

Enable and start all systemd units:
```bash
sudo systemctl enable --now hbank-api.service
sudo systemctl enable --now hbank-payment-plans.timer
```

Set the system timezone to UTC to avoid problems with payment plan execution:
```bash
sudo timedatectl set-timezone UTC
```

You can optionally serve the frontend using this api by simply specifying the `frontendRoot` configuration option. Example:
```jsonc
{
  "frontendRoot": "web"
}
```

## Configuration

HBank-API is looking for configuration in the following locations in order of decreasing precedence: `<working dir>/config.json`, `XDG_CONFIG_HOME/hbank/config.json`.
*Note:* The first config file to be found is used and all others are discarded.

### Default configuration
```jsonc
{
  "debug": false, // !!DO NOT USE IN PRODUCTION!! Disables SameSite for cookies. Returns error messages on HTTP-500 responses.
  "dbVerbose": false, // Prints all sql queries to stdout
  "serverPort": 80, // The port to use for the webserver (if ssl: default = 443)
  "ssl": false, // Enable ssl
  "sslCertPath": "", // Path to ssl cert file
  "sslKeyPath": "", // Path to ssl key file
  "domainName": "hbank.example", // Domain to use for cookies, name of totp, links in email
  "baseURL": "https://hbank.example", // The URL the application is located at
  "emailEnabled": false, // !!ENABLE OR CORE FUNCTIONALITY WILL BE BROKEN!! Send emails
  "emailHost": "", // Host to use for sending emails
  "emailPort": 0, // Port to use for sending emails
  "emailUsername": "", // Username for email account to use for sending emails
  "emailPassword": "", // Password for email account to use for sending emails
  "minNameLength": 3, // Min length of names like usernames, group names, transaction names, payment plan names, etc.
  "maxNameLength": 30, // Max length of names like usernames, group names, transaction names, payment plan names, etc.
  "minDescriptionLength": 0, // Min length of descriptions like group descriptions, transaction descriptions, payment plan descriptions, etc.
  "maxDescriptionLength": 256, // Max length of names like group descriptions, transaction descriptions, payment plan descriptions, etc.
  "maxProfilePictureFileSize": 10000000, // Max size of uploaded profile pictures in bytes
  "maxPageSize": 100, // Max allowed page size for lists
  "frontendRoot": "", // Path to the web root of the frontend
  "idProvider": "" // URL pointing to an OpenID Connect identity provider (must match the issuer value of the provider)
}
```

## License

Copyright © 2021-2022 Julian Hofmann

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

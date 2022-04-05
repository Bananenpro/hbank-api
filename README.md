![H-Bank](assets/banner.png)
<h1 align="center">H-Bank API</h1>

![License](https://img.shields.io/github/license/Bananenpro/hbank-api)
![Commit Activity](https://img.shields.io/github/commit-activity/m/Bananenpro/hbank-api)
![Go Version](https://img.shields.io/github/go-mod/go-version/Bananenpro/hbank-api)
![Total Lines](https://img.shields.io/tokei/lines/github/Bananenpro/hbank-api)
![Issues](https://img.shields.io/github/issues/Bananenpro/hbank-api)
![Stars](https://img.shields.io/github/stars/Bananenpro/hbank-api)
![Forks](https://img.shields.io/github/forks/Bananenpro/hbank-api)

## Description

The API for the [H-Bank](https://github.com/Bananenpro/hbank) application.

## Setup

### Cloning the repo

```bash
git clone https://github.com/Bananenpro/hbank-api.git
cd hbank-api
```

### Installing dependencies

- Go 1.17+
- C compatible compiler such as gcc 4.6+ or clang 3.0+

#### Windows

Using [chocolatey](https://chocolatey.org/) as administrator:

```powershell
choco install golang mingw
```

#### Arch Linux

```bash
sudo pacman -S go gcc
```

### Running

Make sure you are in the root directory of the project.

#### API

The api and webserver application.

```bash
go run ./cmd/hbank-api/main.go
```

#### Payment plans

The service that executes payment plans.

```bash
go run ./cmd/hbank-payment-plans/main.go
```

### Building

#### API

```bash
go build -o bin/hbank-api ./cmd/hbank-api/main.go
```

The resulting binary can be found at `bin/hbank-api`.

#### Payment plans

```bash
go build -o bin/hbank-payment-plans ./cmd/hbank-payment-plans/main.go
```

The resulting binary can be found at `bin/hbank-payment-plans`.

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
  "jwtSecret": "&#aMv5g^m5fzon29eY!QVqkRLMqugFYz",
  "emailEnabled": true,
  "emailHost": "smtp.gmail.com",
  "emailPort": 587,
  "emailUsername": "example@gmail.com",
  "emailPassword": "verysecurepassword",
  "captchaEnabled": true,
  "captchaVerifyUrl": "https://hcaptcha.com/siteverify",
  "captchaSecret": "verysecretcaptchasecret",
  "captchaSitekey": "my-sitekey-bla-bla-1234"
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
  "domainName": "hbank", // Domain to use for cookies, name of totp, links in email
  "jwtSecret": "", // !!SET TO SOMETHING SECURE!! Secret key to use for signing jwt tokens
  "captchaEnabled": false, // Enable CAPTCHA
  "captchaVerifyUrl": "", // URL to verify CAPTCHA response
  "captchaSecret": "", // Secret for CAPTCHA service
  "captchaSiteKey": "", // CAPTCHA sitekey
  "emailEnabled": false, // !!ENABLE OR CORE FUNCTIONALITY WILL BE BROKEN!! Send emails
  "emailHost": "", // Host to use for sending emails
  "emailPort": 0, // Port to use for sending emails
  "emailUsername": "", // Username for email account to use for sending emails
  "emailPassword": "", // Password for email account to use for sending emails
  "minNameLength": 3, // Min length of names like usernames, group names, transaction names, payment plan names, etc.
  "maxNameLength": 30, // Max length of names like usernames, group names, transaction names, payment plan names, etc.
  "minDescriptionLength": 0, // Min length of descriptions like group descriptions, transaction descriptions, payment plan descriptions, etc.
  "maxDescriptionLength": 256, // Max length of names like group descriptions, transaction descriptions, payment plan descriptions, etc.
  "minPasswordLength": 6, // Min password length
  "maxPasswordLength": 64, // Max password length
  "minEmailLength": 3, // Min allowed length of email addresses
  "maxEmailLength": 64, // Max allowed length of email addresses
  "maxProfilePictureFileSize": 10000000, // Max size of uploaded profile pictures in bytes
  "bcryptCost": 10, // Bcrypt cost to use for hashing passwords
  "pbkdf2Iterations": 10000, // Number of iterations of PBKDF2 for hashing tokens
  "recoveryCodeCount": 5, // Number of recovery codes to generate per user
  "loginTokenLifetime": 300, // Time after which login tokens like password or two factor tokens expire in seconds
  "emailCodeLifetime": 300, // Time after which codes sent via email like password reset codes expire in seconds
  "authTokenLifetime": 600, // Time after which auth tokens expire in seconds
  "refreshTokenLifetime": 31536000, // Time after which refresh tokens expire in seconds
  "sendEmailTimeout": 180, // Timeout for sending the same email to the same address in seconds
  "maxPageSize": 100 // Max allowed page size for lists
  "frontendRoot": "" // Path to the web root of the frontend
}
```

## License

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.

## Copyright

Copyright © 2021-2022 Julian Hofmann

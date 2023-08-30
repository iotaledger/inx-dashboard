# inx-dashboard
A node dashboard that uses INX

## Building inx-dashboard

Clone the repository with submodules

```bash 
git clone --recurse-submodules https://github.com/iotaledger/inx-dashboard.git
```

Go to the root directory of the repo and call
```bash
./scripts/build.sh
```

## Setting password for login

To generate a password hash, and it's salt you need to use the <a href=https://github.com/iotaledger/hornet>Hornet</a> docker image.

After installing docker use the hornet tools to generate the password hash:
```bash
docker run -it iotaledger/hornet:2.0-rc tools pwd-hash
```

You can set the hash and the salt with these parameters:
```bash
./inx-dashboard --dashboard.auth.passwordHash YOURHASH --dashboard.auth.passwordSalt YOURSALT
```

Do not forget to change username with ```--dashboard.auth.username```

## Getting full list of parameters

```bash
./inx-dashboard --help --full
```

# inx-dashboard
A node dashboard that uses INX

## Building inx-dashboard

Clone

```bash 
git clone https://github.com/iotaledger/inx-dashboard.git
```

Go to the root directory of the repo and call
```bash
./scripts/build.sh
```

**Or** if that doesn't work because of missing dependencies etc do following steps:

1. ```bash 
    git clone https://github.com/iotaledger/inx-dashboard.git
   ```
2. ```bash
    cd inx-dashboard
   ```
3. ```bash
   git submodule update --init --recursive
   ```
4. ```bash
   cd node-dashboard
   ```
5. ```bash
   npm install && npm run build
   ```
6. ```bash
   cd ..
   ```
7. ```bash
   ./build.sh && go build
   ```

## Setting password for login

To generate a password hash, and it's salt you need to clone and build <a href=https://github.com/iotaledger/hornet>Hornet</a>

After building use hornet tools to generate the password hash:
```bash
./hornet tools pwd-hash
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
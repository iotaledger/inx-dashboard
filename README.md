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
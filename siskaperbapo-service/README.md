# How to Setup this backend

Requirement:
- Go Installed (Idk what version, but mine's Go v1.12.1)
- Git

### For Local Development

1. Clone this repo
```
git clone https://github.com/Majadigi-UB-Kelompok-10/Backend.git
```

2. Run initial go setup
```
go mod
```
or
```
go mod tidy
```

3. Get the .env from your colleague or use dotenvx (still with permission of your colleague)

You will need dotenvx to use the encrypted .env
```
dotenvx decrypt -f .env.encrypted
```
then
```
mv .env.encrypted .env
```

or similar equivalent to rename the file

4. Run the app
```
go run ./cmd/api/main.go
```

5. Or build
```
go build ./cmd/api/main.go
```

### For Container Solution

Create .postgre-data folder if not exist:
```
mkdir .postgre-data
```

Create go build binary if necessary:
```
CGO_ENABLED=0 GOOS=linux go build -o service-siskaperbapo ./cmd/api/main.go
```

Prepare .env:
```
dotenvx decrypt -f .env.docker.encrypted
```
then
```
mv .env.docker.encrypted .env
```

Run docker compose:
```
docker compose up
```
or detached:
```
docker compose up -d
```

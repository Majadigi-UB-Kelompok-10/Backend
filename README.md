### To Decrypt:
1. Put .env.keys in here root directory
2. Run this Command:
```
dotenvx decrypt -fk .env.keys -f bapenda-service/.env.encrypted -f sidita-service/.env.encrypted -f sidita-service/.env.docker.encrypted -f siskaperbapo-service/.env.docker.encrypted -f siskaperbapo-service/.env.encrypted
```
3. Change/copy file to have different filename
```
cp bapenda-service/.env.encrypted bapenda-service/.env /
cp sidita-service/.env.encrypted sidita-service/.env /
cp sidita-service/.env.docker.encrypted sidita-service/.env.docker /
cp siskaperbapo-service/.env.docker.encrypted siskaperbapo-service/.env
```

Above command might change depending on how many encrypted envs

### To Encrypt:

Run this command:
```
dotenvx decrypt -fk .env.keys -f bapenda-service/.env.encrypted -f sidita-service/.env.encrypted -f sidita-service/.env.docker.encrypted -f siskaperbapo-service/.env.docker.encrypted -f siskaperbapo-service/.env.encrypted
```

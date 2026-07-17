# install
```shell
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl ca-certificates
 sudo install -d /usr/share/postgresql-common/pgdg
curl -o /tmp/pgdg.asc --fail https://www.postgresql.org/media/keys/ACCC4CF8.asc
sudo mv /tmp/pgdg.asc /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc

sudo sh -c 'echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] \
https://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main 19" \
> /etc/apt/sources.list.d/pgdg.list'
sudo apt-get update
sudo apt-get install -y postgresql-19

sudo systemctl start postgresql@19-main

```
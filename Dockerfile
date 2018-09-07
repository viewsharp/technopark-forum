FROM ubuntu:18.04

MAINTAINER Vladimir V. Atamanov

# Обвновление списка пакетов
RUN apt-get -y update
# RUN apt-get -y upgrade

#
# Установка postgresql
#
ENV PGVER 10
RUN apt-get install -y postgresql-$PGVER

# Run the rest of the commands as the ``postgres`` user created by the ``postgres-$PGVER`` package when it was ``apt-get installed``
USER postgres

COPY db/ddl.sql ddl.sql

# Create a PostgreSQL role named ``docker`` with ``docker`` as the password and
# then create a database `docker` owned by the ``docker`` role.
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    psql -d docker -f ddl.sql &&\
    /etc/init.d/postgresql stop

# Adjust PostgreSQL configuration so that remote connections to the
# database are possible.
RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf

# And add ``listen_addresses`` to ``/etc/postgresql/$PGVER/main/postgresql.conf``
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf

# Expose the PostgreSQL port
EXPOSE 5432

# Add VOLUMEs to allow backup of config, logs and databases
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

# Back to the root user
USER root

#
# Сборка проекта
#

# Установка golang
RUN apt install -y golang-1.10 git

# Выставляем переменную окружения для сборки проекта
ENV GOROOT /usr/lib/go-1.10
ENV GOPATH /opt/go
ENV PATH $GOROOT/bin:$GOPATH/bin:/usr/local/go/bin:$PATH

# Копируем исходный код в Docker-контейнер
WORKDIR $GOPATH/src/github.com/viewsharp/TexPark_DBMSs/
ADD restapi/ $GOPATH/src/github.com/viewsharp/TexPark_DBMSs/restapi/
ADD db/ $GOPATH/src/github.com/viewsharp/TexPark_DBMSs/db/

# Подтягиваем зависимости
RUN go get github.com/valyala/fasthttp
RUN go get github.com/buaazp/fasthttprouter
RUN go get github.com/lib/pq
RUN go get github.com/mailru/easyjson

# Устанавливаем пакет
RUN go install ./restapi

EXPOSE 5000

CMD service postgresql start &&\
    restapi 5000

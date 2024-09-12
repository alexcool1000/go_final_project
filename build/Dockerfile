FROM ubuntu:latest

WORKDIR /todolist

COPY ./build/todolist ./

COPY ./web ./web

ENV TODO_PORT="7540" TODO_DBFILE="./scheduler.db" TODO_PASSWORD="123" TODO_WEBDIR="./web"

EXPOSE 7540

CMD [ "./todolist" ]
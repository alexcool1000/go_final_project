FROM ubuntu:latest

WORKDIR /todolist

COPY ./todolist ./

COPY ./web ./web

ENV TODO_PORT="7540" TODO_DBFILE="./scheduler.db" TODO_PASSWORD="123"

EXPOSE 7540

CMD [ "./todolist" ]
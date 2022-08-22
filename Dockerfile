FROM nginx:latest as server
COPY ./public /usr/share/nginx/html


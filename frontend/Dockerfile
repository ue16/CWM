FROM node:24 as build
WORKDIR /frontend
COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile
COPY . .
RUN yarn build


FROM caddy:2
WORKDIR /frontend
COPY --from=build /frontend/dist ./

EXPOSE 80
CMD [ "caddy", "file-server", "--root", "/frontend" ]

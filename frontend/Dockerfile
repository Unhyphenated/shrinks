FROM node:20-alpine AS build

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

ARG VITE_SHORT_DOMAIN
ARG VITE_API_URL

ENV VITE_SHORT_DOMAIN=$VITE_SHORT_DOMAIN
ENV VITE_API_URL=$VITE_API_URL

# This creates the /dist folder
RUN npm run build

FROM nginx:stable-alpine

# Copy the static files from the build stage to Nginx
COPY --from=build /app/dist /usr/share/nginx/html

COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80 for Nginx
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
# syntax=docker/dockerfile:1

# ---------- Frontend build ----------
FROM node:22 AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend ./

ARG VITE_API_URL
ENV VITE_API_URL=${VITE_API_URL}

RUN npm run build

# ---------- Backend build ----------
FROM golang:1.24 AS backend
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend .
COPY --from=frontend /app/frontend/dist ./frontend/dist

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./main.go

# ---------- Final image ----------
FROM debian:bullseye-slim
WORKDIR /app
COPY --from=backend /app/backend/main .
COPY --from=backend /app/backend/frontend/dist ./frontend/dist

EXPOSE 8080
CMD ["./main"]

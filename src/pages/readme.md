# Football Manager Player Parser

A tool for parsing and displaying Football Manager HTML player data with a Vue.js + Quasar UI interface and Go backend.

## Project Setup

### Backend (Go)

```bash
# Initialize Go module
go mod init fm24golang

# Install dependencies
go get golang.org/x/net/html
go mod tidy
```

### Frontend (Vue + Quasar)

```bash
# Install Node.js dependencies
npm install
```

## Running the Application

### Start the Backend API

```bash
# Start the Go backend API
go run main.go
```

The backend will run on http://localhost:8080.

### Start the Frontend Development Server

```bash
# Start the Vue.js frontend with hot-reload
npm run dev
```

The frontend will run on http://localhost:3000.

## How to Use

1. Start both the backend API and frontend development server
2. Access the frontend at http://localhost:3000
3. Select an HTML file exported from Football Manager containing player data
4. Click "Upload and Parse" to process the file
5. Use search, sort, and pagination to explore the data

## Building for Production

```bash
# Build the frontend for production
npm run build
```

This will generate static files in the `dist` directory that can be served by any static file server.


# Querying LLM:

Use this -> codeweaver --ignore "node_modules,\.git"
or ./CodeWeaver --ignore "node_modules,testdata.html,.git,CodeWeaver"

# Immage Filters in Go

This project implements **image processing filters in Go**, both **sequential** and **parallel**, using **goroutines**, a **TCP client–server architecture**, and several **performance analysis programs**.

The goal is to study parallelism, scalability, and performance trade-offs in Go.
This project was made by Farah Gattoufi, Anas Sfar and Yousra Mounim.
---

## Project Structure

```
.
├── filters/
│   ├── seq.go          # Sequential filter implementations
│   └── parallel.go     # Parallel implementations (goroutines + workers)
│
├── TCP/
│   ├── server.go       # TCP server
│   └── client.go       # TCP client
│
├── performance/
│   ├── image_size.go        # Impact of image size on execution time
│   ├── seq_vs_parallel.go   # Sequential vs parallel comparison
│   └── scaling_workers.go   # Worker scalability analysis
│
└── README.md
```

---

## Available Filters

- `grayscale` – grayscale conversion  
- `invert` – color inversion  
- `blur` – box blur  
- `gaussian` – gaussian blur  
- `sobel` – edge detection  
- `median` – median filter  
- `pixelate` – mosaic effect  
- `oilpaint` – oil painting effect  

---
Important : The image needs to be in the same file as the scripts you want to run.
---

## TCP Client–Server Mode

### Run the server

```bash
go run server.go parallel.go
```

- Applies filters in parallel
- Allows or automatically selects the number of workers
- Measures filter execution time

---

### Run the client

```bash
go run client.go
```

The client:
- Displays filter descriptions
- Saves the output in the same format
- Displays server-side execution time

---

## Performance Analysis

### Image size impact

```bash
go run image_size.go parallel.go seq.go
```

### Sequential vs Parallel

```bash
go run seq_vs_parallel.go parallel.go seq.go
```

### Worker scalability

```bash
go run scaling_workers.go parallel.go seq.go
```

---

## Technologies

- Go (Golang)
- Goroutines
- TCP networking
- Image processing
- Performance measurement

---

## Purpose

Academic project exploring parallelism, performance analysis, and TCP communication in Go.

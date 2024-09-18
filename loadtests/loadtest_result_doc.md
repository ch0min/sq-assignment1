# Load Testing Report

## Overview

This document provides the results and analysis of the load testing performed on the `/api/todos` endpoint of our application. The testing was conducted using k6 to evaluate the API's performance under a simulated load of 50 concurrent users.

## Test Details

- **Tool Used:** k6
- **Test Script Location:** `./loadtests/loadtest.js`
- **Execution Date:** 18-09-2024

## Test Configuration

- **Scenario:** Simulate up to 50 concurrent users
- **Duration:** 4 minutes
- **Endpoints Tested:** `/api/todos` (GET request)

## Results

### Summary

- **Total Requests:** 8959
- **Success Rate:** 100% (0 failed requests)
- **Average Response Time:** 3.19ms
- **Peak Response Time:** 433.48ms
- **Throughput:** 37.24 requests per second
- **Virtual Users:** Up to 50

### Detailed Metrics

- **Data Received:** 37 MB
- **Data Sent:** 797 kB
- **Average Request Duration:** 3.19ms
- **Average Connection Time:** 2.5µs
- **Average Waiting Time:** 3.16ms
- **Requests Per Second:** 37.24
- **Iterations:** 8959

### Key Performance Indicators

- **HTTP Requests Blocked:** Average 5.08µs
- **HTTP Requests Connecting:** Average 2.5µs
- **HTTP Requests Receiving:** Average 30.42µs
- **HTTP Requests Sending:** Average 971ns
- **HTTP Requests TLS Handshaking:** Average 0s
- **HTTP Requests Waiting:** Average 3.16ms

## Analysis

- The API handled the load well with a high success rate and an average response time of 3.19ms.
- The maximum response time of 433.48ms indicates that while most requests are processed quickly, there were occasional delays.
- The throughput of 37.24 requests per second was consistent, showing stable performance under the simulated load.

---

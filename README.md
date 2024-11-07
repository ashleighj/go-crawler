# webcrawler

*Author: Ashleigh Waldek, 2024*

## Introduction

*webcrawler* is a web crawler! Surprise!

## Important Characteristics

- ***Parallelisation*** for efficiency and scalability
- ***Robustness*** to handle edge cases like bad HTML, unresponsive servers, malicious links, etc.
- ***Politeness*** so as not to inundate target pages with too many/frequests subsequent requests
- ***Performance*** - breadth-first search used (usually a better choice than depth-first for web crawlers as the depth can be very deep; opportunity for more parallel goroutines to be started early)
- ***Efficiency*** - not crawling previously seen links or content

## Design

<img width="865" alt="Screenshot 2024-11-07 at 13 56 02" src="https://github.com/user-attachments/assets/801ce257-a33f-4c01-ba51-099663882d2c">

###
- filtering  - 1 goroutine
- routing    - 1 goroutine
- pre-crawl  - goroutine per unique host in filtered URLs
- crawl      - goroutine per host URL visited

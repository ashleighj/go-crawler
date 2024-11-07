# webcrawler

*Author: Ashleigh Waldek*

## Introduction

*webcrawler* is a web crawler! Surprise!

## Important Characteristics

- *Parallelisation* for efficiency and scalability
- *Robustness* to handle edge cases like bad HTML, unresponsive servers, malicious links, etc.
- *Politeness* so as not to inundate target pages with too many/frequests subsequent requests
- *Performance* - breadth-first search used (usually a better choice than depth-first for web crawlers as the depth can be very deep; opportunity for more parallel goroutines to be started early)
- *Efficiency* - not crawling previously seen links or content

## Design


[[ seed URLS ]] 
    ^       --> filtering 
    |                 --> [[ filtered URLS ]] 
    |                             --> routing (channel/queue per host)
    |                                                         --> [[ routed URLS  ]] 
    |                                                                      --> GETing and link extraction |  
    |                                                                                                     |
    |                                                                                                     v 
    -------------<-------------------<-------------------<--------------------<---------------<------------

filtering  - 1 goroutine
routing    - 1 goroutine
pre-crawl  - goroutine per unique host in filtered URLs
crawl      - goroutine per host URL visited

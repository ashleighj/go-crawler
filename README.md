# webcrawler

*Author: Ashleigh Waldek*

## Introduction

*webcrawler* is a web crawler! Surprise!

## Important Characteristics

- *Parallelisation* for efficiency and scalability
- *Robustness* to handle edge cases like bad HTML, unresponsive servers, malicious links, etc.
- *Politeness* so as not to inundate target pages with too many/frequests subsequent requests
- *Performance* - breadth-first search used (usually a better choice than depth-first for web crawlers as the depth can be very deep; opportunity for more parallel goroutines to be started early)

## Design




## Possible improvements, given more time

- add functionality and another layer of channeling that would allow for prioritising URLs to be downloaded
- add external storage
- improve logging; add better formatting, more information, masking etc


## How this could be scaled
# go-challenge

## Design Decisions

* Use of context package with timeout to make sure the application finishes its job in 500 ms
* The context timeout is set to 490 ms to give it a buffer of 10 ms to encode and return response
* Each of the url is validated to check if it is a valid url or not
* The valid urls are processed in independent subroutine - task
* Another subroutine - sorter is also running which sorts and merges the number
* As soon as the response is retrieved, the result is put into a channel which is being listened by the sorter subroutine
* When the sorter subroutine gets a list of numbers, it checks if it is already present in the result set. If not the number is appended to the result set
* After all the number fetched in a single channel message is processed, it is sorted using the go sort package.
* The sorted result is then pushed to a channel which is being listened in the main program
* In the main program, as soon as a number list appears in the result channel, it is stored in a local variable
* Once all the urls are processed or time out is reached (whichever is earlier), the most recent result is returned in the response body
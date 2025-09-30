# Instructions

To run the solution, navigate to the directory where it's located and type `go run main.go`.  The server will start up and start listening on port 8080.  Press Ctrl-C to stop it.

Once the server is running, run the device simulator from its directory by typing `./device-simulator-linux-amd64 -port 8080`.  It will output the `results.txt` file to the same directory.

## Example
Assume the solution and the device simulator were both downloaded to the `Downloads` directory.  Open two terminal windows.

In the first one type:
```
cd ~/Downloads/fleetsy
go run main.go
```

In the second one type:
```
cd ~/Downloads
./device-simulator-linux-amd64 -port 8080
```

Once the device simulator has finished running, it will output the results to the screen and to a `results.txt` file in the `~/Downloads` directory.

# Writeup

## How long did you spend working on the problem?  What did you find to be the most difficult part?

I am not sure how long I spent working on the problem.  I was out of town almost all weekend and had to fit it in where I could without disrupting my evening classes.  It was difficult to get any real momentum going as a result.  At as guess I can say I spent at least six hours on it.  There are many optimizations I would perform if I had more time.

I found the problem had a few challenges for me.  Go is not a language that I work with on a daily basis so I found myself slowed down by being a little rusty at some aspects.  On top of that I don't think I've ever stood up a brand new API from scratch before.  All the previous Go projects I've worked on were existing ones that needed bug fixes and new features.  The fundamental pieces were already done.  The chi router I ended up using also gave me a bit of trouble trying to get it to work properly since that was the first time using it.  If I had to do it again I would do a bit more research on the different options and pick one that would be easier to work with.

## How would you modify your data model or code to account for more kinds of devices?

The data model I used here is obviously only suitable for such a small project.  Using two maps with a single mutex for safe operation isn't a very scalable solution.  It would quickly run into issues with requests waiting on mutex locks in order to perform their operations and would drastically slow down the overall performance of the system.  However, the mutex was necessary to ensure safe access to the device maps across multiple threads.  I deemed it to be an acceptable compromise for the size of this project.  The structs I designed to hold the data internally worked well for this use case but I would need to take a more considered approach to be able to handle more types of devices.  As it is they could be reworked and consolidated to simplify things internally.

For an actual production system I would use a proper database solution.  Which one depends on how many devices would be in use, how much data would be required for each one, and how much load the database would be expected.  Sqlite would be an appropriate solution if it could be guaranteed that the load on the database would not be more than it could handle, such as a small home automation solution.  For more large scale use cases, I would choose Postgres due to it's high performance and ability to scale to very large workloads.  With a database backing the data, I would spend more time designing structs to better fit the devices.  Due to time constraints I ended up creating a struct for each incoming message type which is an unsustainable solution as the number of devices grows.  Using a key-value store like redis for caching would also significantly improve access times to the data in a production environment.

## Discuss your solution's runtime complexity

I tried to keep my runtime complexity as close to linear, or O(n) where n is the number of data points for a specific device, as I could.  Since I used a map for data storage, any access operations are O(1) to retrieve the reference to the data slice.  The intention here was to emulate either a database query or a key-value store, both of which would provide quick, direct access to the data in question without needing to iterate through all the different devices.

POST operations are O(1) in the best case due to how Go slices work behind the scenes and O(n) in the worst case if the underlying array is full.  If the underlying array is full, the Go runtime will create a new, larger array to store the data for the slice and then copy all the data from the old array to the new one.  This allows slice operations to be very fast in some cases but at the expense of using up some additional memory.  For this problem I felt that was an acceptable trade off given the amount of data used.  So the worse case complexity here is O(n)

GET operations are also O(1) to retrieve the data for each device, and then O(n) to perform the necessary calculations due to the requirement to calculate the average upload time.  Since it was assumed that heartbeat messages were received in chronological order, the worst case complexity there is still O(1) due to the access pattern implemented.  The complexity would obviously increase if this was not the case.  Therefore the worst case complexity here is O(n).

Since all operations are designed to be O(n) in the worst case, the overall worst case complexity occurs when all devices are queried for their stats at once.  This would result in O(m * n) runtime where m is the number of devices and n is the number of datapoints for each device.  This access pattern is deemed to be unlikely in the scope of this problem but could very well happen in a production environment.  The use of caching and store procedures could help improve real world performance in production, but ultimately the calculations need to be done for each device and each data point, so it remains O(m * n)

# Improvements

There are a number of improvements that could be made here since this solution was rushed due to time constraints.  The most obvious is that there are no automated tests included which would provide reassurance that all functions and operations are behaving as expected.  In addition, automated testing helps ensure there are no regression bugs when working on a large project with ongoing development.

Code reuse is another area that could be improved upon.  There are several instances where I simply copied and pasted code from one function to another.  Again, if proper utility functions were created this would cut down on code reuse and make future maintenance easier.

I originally anticipated running this in docker, but ran out of time.  So the Dockerfile can still be found in the git history but has been removed from the final submission as a result.  Again, docker would be very useful in production as it gives a consistend runtime for the binary across multiple platforms and go-based docker services can be very, very small.

For a production use case, it is probable that the uptime would not be calculated for every data point in the device's history, but instead for a specific period of time, such as daily or hourly.  With well-defined access patterns for this data, I could design an in-memory table or key-value store to keep the relevant data easily accessed to avoid consuming too much memory with redundant data.  In addition, if we can make the assumption that we will never be receiving new data for previous time periods after a defined cut-off, we could store the final calculated statistics in another table to make retrieval even faster.
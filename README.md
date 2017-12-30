# aptblueprint

This application grabs data from the free data at ourairports.com, sticks it in a local database, and generates a blueprint-style diagram of the airport's runways.  Optionally, there can get posted to twitter if the necessary tokens are in the config file.  For data management, this uses my own in-development [airport data library](http://github.com/kaosfere/aptdata).

## todos

* add an option for specifying the minimum number of runways
* similarly, paramterize output picture size
* skip file generation if in post mode, and just base64 the raw image data?
* test suite!

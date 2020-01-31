# osp
outsource project for test

Service requires config.toml to start (exists in project folder). There is already prebuilt binary for ubuntu x64. 

Images are taken from internet only for testing. 

Service consists of two parts: http web server (:8080 by default) and image processor.
Http serves as UI and passes requests to image processor.
Image processor is an interface with single method - ProcessImage. Default implementation uses chan to 
communicate with goroutine that makes processing.

Of course, not all validations are made, not all errors are handled - it's just demo. I tried to show common 
practices and concepts (interfaces, working with goroutines, defering i/o on channels with timers, different types 
of error handling, defers, etc).

Repo contains some tests that cover difficult places 
(some tests are complicated - for example setuping temp dirs to test processor and check for specific errors)

Why vanilla JS? I dont' like to add unnecessary complexity in places where it is not required. Html page is very little 
and adding react or angular would be overkill to my mind. Hope you agree with me.

What to improve:
- add more tests for complex cases with already (existing files, etc)
- replace config toml with env or flags (parsing toml and debugging is little quicker than other variants)
- use some package for code generation (now html template is build in manually just to make building project easier)
- 2nd and 3rd steps will allow using just single executable without additional files at all (config and page template)

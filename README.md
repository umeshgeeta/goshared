# goshared

This repository contains various go packages which any developer can import and consume. 

1) util package contains simple wrapper functions I found useful around Logging and Configuration management. These wrapper functions are around Lumberjack Rotating Logging and Cleanenv configuration management package from ilyakaznacheev. Initial tagged version is ready for consumption.

2) executor package is bit ambitious. I come from Java Object Oriented Serverside coding background. Like most programmers, there are many things I like about Java and I miss those anytime I pickup to learn a new programming language. Java scheduler and thread pool executor framework is one such feature I like in Java. In this package I am working on developing a similar mechanism. As per experts, Go Lang has strong Concurrency Support. But it's philosophy is to remain 'bare metal'. Besides it is a young language compared to Java. I needed 'executor type functionality' in another of my copy-right protected project. I am sure Go Community would have many solid packages to address this need. But I wanted to take it as an exercise so as I understand Go Lang Concurrency as well as develop a building block exactly what I needed in another copy-right protected project. Hence the efforts.

Package 'executor' is not ready for consumption yet. I am in the process of implementing various tests to validate the functionality as intended. Design is fairly clear. I will update as soon as it is ready for the consumption.



Feel free to reach out to me (umesh409@yahoo.com) for:

- how to use the code mentioned here, or
- if you need any additional functionality in the current code. I will try my best to have it included.

Code is under MIT License, so it is least restrictive. (Please stick to 'master' branch which has working code while code in other branches is still under development and not ready for consumption.)

Have a look and enjoy! 

Thanks for visiting!!!

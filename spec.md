NewLatinLinker
==================

This is the new, improved version of LatinLinker. The old version was done before I had access to a good web server, so I had to 
do everything client-side, which wasn't very good in a lot of ways. This new version is actually simpler, not more complex, but that makes it better. 

My new plan is to generate static HTML pages with a script that do the same things as the old, dynamically generated pages did, but they run a lot faster
and can be more complete because I'm not limited by same-origin.

More specifically, I'm going to write a python script that just scrapes LatinLibrary and recreates the file structure. The only changes it will make are making all the
words links to dictionary definitions, as the old one did. Then I'm going to gzip all those documents, put them on the server, and be done. It will be super simple and fast. 

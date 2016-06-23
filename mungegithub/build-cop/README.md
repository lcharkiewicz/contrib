# Build Cop tools

This directory contains tools for build cops. They give an easier access to the
merge queue's admin console.

**They require you to have kubectl installed and pointing at the cluster running
the submit queue.** 

## resolved.sh

Use `resolved.sh` to mark a test as manually resolved (or not), so the merge
queue will merge in spite of a failure. Arguments are the job name and run number.

## emergency.sh

Run `emengency.sh` to halt merges. Run `emergency.sh resume` to start them
again.
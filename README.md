# go_sync_pool
Improvement over golang sync.Pool

Improvements include:
1. Controlled cleaning of the pool. No need to clean the pool on every GC. Use EWMA to find expected number of items in the pool and cleanup accordingly.
2. Facility to drain and close the pool when done using it.
3. Ability to choose from (1) lock-free queue (2) lock-free stack OR (3) golang channel as a background implementation.
4. Ability to set max size of the pool to control total memory usage.

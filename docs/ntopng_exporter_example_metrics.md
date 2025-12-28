# List of Metrics Emitted by ntopng-exporter

<!-- markdownlint-disable -->
All metrics prefixed with `go_` indicate application performance metrics from ntopng-exporter itself.

All metrics having to do with ntopng are prefixed with `ntopng_`. These are the current subsets of metrics:
- `ntopng_interface_` metrics - These metrics are all labeled with the interface name and the interface ID that ntopng keeps internally. They indicate metrics that are specific to an individual interface
- `ntopng_host_` metrics - These metrics are all labeled with the IP, MAC address, interface name, interface ID, and name of the host (if ntopng can find it). They indicate metrics that are specific to individual hosts on a given interface.

```
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary

# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge

# HELP go_info Information about the Go environment.
# TYPE go_info gauge

# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
# TYPE go_memstats_alloc_bytes gauge

# HELP go_memstats_alloc_bytes_total Total number of bytes allocated, even if freed.
# TYPE go_memstats_alloc_bytes_total counter

# HELP go_memstats_buck_hash_sys_bytes Number of bytes used by the profiling bucket hash table.
# TYPE go_memstats_buck_hash_sys_bytes gauge

# HELP go_memstats_frees_total Total number of frees.
# TYPE go_memstats_frees_total counter

# HELP go_memstats_gc_cpu_fraction The fraction of this program's available CPU time used by the GC since the program started.
# TYPE go_memstats_gc_cpu_fraction gauge

# HELP go_memstats_gc_sys_bytes Number of bytes used for garbage collection system metadata.
# TYPE go_memstats_gc_sys_bytes gauge

# HELP go_memstats_heap_alloc_bytes Number of heap bytes allocated and still in use.
# TYPE go_memstats_heap_alloc_bytes gauge

# HELP go_memstats_heap_idle_bytes Number of heap bytes waiting to be used.
# TYPE go_memstats_heap_idle_bytes gauge

# HELP go_memstats_heap_inuse_bytes Number of heap bytes that are in use.
# TYPE go_memstats_heap_inuse_bytes gauge

# HELP go_memstats_heap_objects Number of allocated objects.
# TYPE go_memstats_heap_objects gauge

# HELP go_memstats_heap_released_bytes Number of heap bytes released to OS.
# TYPE go_memstats_heap_released_bytes gauge

# HELP go_memstats_heap_sys_bytes Number of heap bytes obtained from system.
# TYPE go_memstats_heap_sys_bytes gauge

# HELP go_memstats_last_gc_time_seconds Number of seconds since 1970 of last garbage collection.
# TYPE go_memstats_last_gc_time_seconds gauge

# HELP go_memstats_lookups_total Total number of pointer lookups.
# TYPE go_memstats_lookups_total counter

# HELP go_memstats_mallocs_total Total number of mallocs.
# TYPE go_memstats_mallocs_total counter

# HELP go_memstats_mcache_inuse_bytes Number of bytes in use by mcache structures.
# TYPE go_memstats_mcache_inuse_bytes gauge

# HELP go_memstats_mcache_sys_bytes Number of bytes used for mcache structures obtained from system.
# TYPE go_memstats_mcache_sys_bytes gauge

# HELP go_memstats_mspan_inuse_bytes Number of bytes in use by mspan structures.
# TYPE go_memstats_mspan_inuse_bytes gauge

# HELP go_memstats_mspan_sys_bytes Number of bytes used for mspan structures obtained from system.
# TYPE go_memstats_mspan_sys_bytes gauge

# HELP go_memstats_next_gc_bytes Number of heap bytes when next garbage collection will take place.
# TYPE go_memstats_next_gc_bytes gauge

# HELP go_memstats_other_sys_bytes Number of bytes used for other system allocations.
# TYPE go_memstats_other_sys_bytes gauge

# HELP go_memstats_stack_inuse_bytes Number of bytes in use by the stack allocator.
# TYPE go_memstats_stack_inuse_bytes gauge

# HELP go_memstats_stack_sys_bytes Number of bytes obtained from system for stack allocator.
# TYPE go_memstats_stack_sys_bytes gauge

# HELP go_memstats_sys_bytes Number of bytes obtained from system.
# TYPE go_memstats_sys_bytes gauge

# HELP go_threads Number of OS threads created.
# TYPE go_threads gauge

# HELP ntopng_host_active_client_flows current number of active client flows for host
# TYPE ntopng_host_active_client_flows gauge

# HELP ntopng_host_active_server_flows current number of active server flows for host
# TYPE ntopng_host_active_server_flows gauge

# HELP ntopng_host_bytes_rcvd number of bytes received for host
# TYPE ntopng_host_bytes_rcvd counter

# HELP ntopng_host_bytes_sent number of bytes sent for host
# TYPE ntopng_host_bytes_sent counter

# HELP ntopng_host_dns_queries_by_type total number of DNS queries by record type
# TYPE ntopng_host_dns_queries_by_type counter

# HELP ntopng_host_num_alerts number of alerts for host
# TYPE ntopng_host_num_alerts gauge

# HELP ntopng_host_packets_rcvd number of packets received for host
# TYPE ntopng_host_packets_rcvd counter

# HELP ntopng_host_packets_sent number of packets sent for host
# TYPE ntopng_host_packets_sent counter

# HELP ntopng_host_total_alerts total number of alerts for host
# TYPE ntopng_host_total_alerts counter

# HELP ntopng_host_total_client_flows total number of client flows for host
# TYPE ntopng_host_total_client_flows counter

# HELP ntopng_host_total_dns_queries total number of DNS queries for host
# TYPE ntopng_host_total_dns_queries counter

# HELP ntopng_host_total_dns_replies total number of DNS replies for host by status
# TYPE ntopng_host_total_dns_replies counter

# HELP ntopng_host_total_server_flows total number of server flows for host
# TYPE ntopng_host_total_server_flows counter

# HELP ntopng_interface_alerted_error_flows current number of alerted error flows
# TYPE ntopng_interface_alerted_error_flows gauge

# HELP ntopng_interface_alerted_flows current number of alerted flows client flows
# TYPE ntopng_interface_alerted_flows gauge

# HELP ntopng_interface_alerted_notice_flows current number of alerted notice flows
# TYPE ntopng_interface_alerted_notice_flows gauge

# HELP ntopng_interface_alerted_warning_flows current number of alerted warning flows
# TYPE ntopng_interface_alerted_warning_flows gauge

# HELP ntopng_interface_bytes_rcvd total number of bytes received
# TYPE ntopng_interface_bytes_rcvd counter

# HELP ntopng_interface_bytes_sent total number of bytes sent
# TYPE ntopng_interface_bytes_sent counter

# HELP ntopng_interface_current_throughput_bps current throughput by direction in bytes per second
# TYPE ntopng_interface_current_throughput_bps gauge

# HELP ntopng_interface_current_throughput_pps current throughput by direction in packets per second
# TYPE ntopng_interface_current_throughput_pps gauge

# HELP ntopng_interface_drops number of drops
# TYPE ntopng_interface_drops counter

# HELP ntopng_interface_num_devices number of devices
# TYPE ntopng_interface_num_devices gauge

# HELP ntopng_interface_num_hosts number of hosts
# TYPE ntopng_interface_num_hosts gauge

# HELP ntopng_interface_num_local_hosts number of hosts on the local network
# TYPE ntopng_interface_num_local_hosts gauge

# HELP ntopng_interface_packets_rcvd total number of packets received
# TYPE ntopng_interface_packets_rcvd counter

# HELP ntopng_interface_packets_sent total number of packets sent
# TYPE ntopng_interface_packets_sent counter

# HELP ntopng_interface_speed current speed of interface in Mbps
# TYPE ntopng_interface_speed gauge

# HELP ntopng_interface_tcp_packet_stats tcp packet stats by type
# TYPE ntopng_interface_tcp_packet_stats counter

# HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE process_cpu_seconds_total counter

# HELP process_max_fds Maximum number of open file descriptors.
# TYPE process_max_fds gauge

# HELP process_open_fds Number of open file descriptors.
# TYPE process_open_fds gauge

# HELP process_resident_memory_bytes Resident memory size in bytes.
# TYPE process_resident_memory_bytes gauge

# HELP process_start_time_seconds Start time of the process since unix epoch in seconds.
# TYPE process_start_time_seconds gauge

# HELP process_virtual_memory_bytes Virtual memory size in bytes.
# TYPE process_virtual_memory_bytes gauge

# HELP process_virtual_memory_max_bytes Maximum amount of virtual memory available in bytes.
# TYPE process_virtual_memory_max_bytes gauge

# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge

# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
```
<!-- markdownlint-restore -->

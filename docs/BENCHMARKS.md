Benchmarks
---

Here you'll find some benchmarks of the different components of the Lura framework in several scenarios.

# Proxy components

## Proxy middleware stack

    BenchmarkProxyStack_single-8                   500000          9106 ns/op        1848 B/op         35 allocs/op
    BenchmarkProxyStack_multi/with_1_backends-8    500000          9183 ns/op        1848 B/op         35 allocs/op
    BenchmarkProxyStack_multi/with_2_backends-8    300000         16130 ns/op        3520 B/op         73 allocs/op
    BenchmarkProxyStack_multi/with_3_backends-8    200000         20780 ns/op        5097 B/op        105 allocs/op
    BenchmarkProxyStack_multi/with_4_backends-8    200000         22420 ns/op        6641 B/op        137 allocs/op
    BenchmarkProxyStack_multi/with_5_backends-8    200000         23966 ns/op        8218 B/op        169 allocs/op

## Proxy middlewares

    BenchmarkNewLoadBalancedMiddleware-8                10000000           435 ns/op         328 B/op          6 allocs/op
    BenchmarkNewConcurrentMiddleware_singleNext-8         500000          9351 ns/op        1072 B/op         18 allocs/op
    BenchmarkNewRequestBuilderMiddleware-8              30000000           115 ns/op         160 B/op          2 allocs/op
    BenchmarkNewMergeDataMiddleware/with_2_parts-8       1000000          6746 ns/op        1360 B/op         20 allocs/op
    BenchmarkNewMergeDataMiddleware/with_3_parts-8        500000         10179 ns/op        1488 B/op         22 allocs/op
    BenchmarkNewMergeDataMiddleware/with_4_parts-8        500000         10299 ns/op        1584 B/op         24 allocs/op

# Response manipulation

## Response property whitelisting

    BenchmarkEntityFormatter_whitelistingFilter/with_0_elements_with_0_extra_fields-8           50000000            80.6 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_1_elements_with_0_extra_fields-8           10000000           441 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_2_elements_with_0_extra_fields-8           10000000           474 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_3_elements_with_0_extra_fields-8           10000000           516 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_4_elements_with_0_extra_fields-8           10000000           519 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_0_elements_with_5_extra_fields-8           50000000            84.3 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_1_elements_with_5_extra_fields-8           10000000           565 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_2_elements_with_5_extra_fields-8           10000000           601 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_3_elements_with_5_extra_fields-8           10000000           638 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_4_elements_with_5_extra_fields-8           10000000           627 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_0_elements_with_10_extra_fields-8          50000000            80.7 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_1_elements_with_10_extra_fields-8          10000000           703 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_2_elements_with_10_extra_fields-8           5000000           746 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_3_elements_with_10_extra_fields-8           5000000           779 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_4_elements_with_10_extra_fields-8           5000000           785 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_0_elements_with_15_extra_fields-8          50000000            81.4 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_1_elements_with_15_extra_fields-8           5000000           845 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_2_elements_with_15_extra_fields-8           5000000           886 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_3_elements_with_15_extra_fields-8           5000000           919 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_4_elements_with_15_extra_fields-8           5000000           929 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_0_elements_with_20_extra_fields-8          50000000            80.9 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_1_elements_with_20_extra_fields-8           5000000           988 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_2_elements_with_20_extra_fields-8           5000000           984 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_3_elements_with_20_extra_fields-8           5000000           998 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_4_elements_with_20_extra_fields-8           5000000          1014 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_0_elements_with_25_extra_fields-8          50000000            78.1 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_1_elements_with_25_extra_fields-8           5000000          1149 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_2_elements_with_25_extra_fields-8           3000000          1279 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_3_elements_with_25_extra_fields-8           3000000          1348 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_4_elements_with_25_extra_fields-8           3000000          1349 ns/op         384 B/op          3 allocs/op

## Response property blacklisting

    BenchmarkEntityFormatter_blacklistingFilter/with_0_elements_with_0_extra_fields-8           50000000            82.4 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_1_elements_with_0_extra_fields-8           30000000           174 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_2_elements_with_0_extra_fields-8           20000000           205 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_3_elements_with_0_extra_fields-8           100000000           63.5 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_4_elements_with_0_extra_fields-8           100000000           62.9 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_0_elements_with_5_extra_fields-8           50000000            80.5 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_1_elements_with_5_extra_fields-8           30000000           175 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_2_elements_with_5_extra_fields-8           20000000           207 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_3_elements_with_5_extra_fields-8           20000000           255 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_4_elements_with_5_extra_fields-8           20000000           299 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_0_elements_with_10_extra_fields-8          50000000            82.9 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_1_elements_with_10_extra_fields-8          30000000           162 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_2_elements_with_10_extra_fields-8          20000000           193 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_3_elements_with_10_extra_fields-8          20000000           229 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_4_elements_with_10_extra_fields-8          20000000           272 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_0_elements_with_15_extra_fields-8          50000000            76.7 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_1_elements_with_15_extra_fields-8          30000000           161 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_2_elements_with_15_extra_fields-8          20000000           195 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_3_elements_with_15_extra_fields-8          20000000           243 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_4_elements_with_15_extra_fields-8          20000000           292 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_0_elements_with_20_extra_fields-8          50000000            81.4 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_1_elements_with_20_extra_fields-8          30000000           161 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_2_elements_with_20_extra_fields-8          20000000           197 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_3_elements_with_20_extra_fields-8          20000000           239 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_4_elements_with_20_extra_fields-8          20000000           289 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_0_elements_with_25_extra_fields-8          50000000            80.9 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_1_elements_with_25_extra_fields-8          30000000           176 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_2_elements_with_25_extra_fields-8          20000000           200 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_3_elements_with_25_extra_fields-8          20000000           250 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_4_elements_with_25_extra_fields-8          20000000           312 ns/op          48 B/op          1 allocs/op

## Response property grouping

    BenchmarkEntityFormatter_grouping/with_0_elements-8             20000000           277 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_grouping/with_5_elements-8             20000000           299 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_grouping/with_10_elements-8            20000000           300 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_grouping/with_15_elements-8            20000000           298 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_grouping/with_20_elements-8            20000000           298 ns/op         384 B/op          3 allocs/op
    BenchmarkEntityFormatter_grouping/with_25_elements-8            20000000           298 ns/op         384 B/op          3 allocs/op

## Repsonse property mapping

    BenchmarkEntityFormatter_mapping/with_0_elements_with_0_extra_fields-8          100000000           61.1 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_1_elements_with_0_extra_fields-8          100000000           63.5 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_2_elements_with_0_extra_fields-8          100000000           61.8 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_3_elements_with_0_extra_fields-8          100000000           63.9 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_4_elements_with_0_extra_fields-8          100000000           63.7 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_5_elements_with_0_extra_fields-8          100000000           64.0 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_0_elements_with_5_extra_fields-8          50000000            81.4 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_1_elements_with_5_extra_fields-8          20000000           177 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_2_elements_with_5_extra_fields-8          20000000           204 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_3_elements_with_5_extra_fields-8          20000000           233 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_4_elements_with_5_extra_fields-8          20000000           266 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_5_elements_with_5_extra_fields-8          20000000           295 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_0_elements_with_10_extra_fields-8         50000000            77.4 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_1_elements_with_10_extra_fields-8         30000000           163 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_2_elements_with_10_extra_fields-8         20000000           198 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_3_elements_with_10_extra_fields-8         20000000           237 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_4_elements_with_10_extra_fields-8         20000000           298 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_5_elements_with_10_extra_fields-8         20000000           331 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_0_elements_with_15_extra_fields-8         50000000            79.5 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_1_elements_with_15_extra_fields-8         30000000           171 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_2_elements_with_15_extra_fields-8         20000000           212 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_3_elements_with_15_extra_fields-8         20000000           265 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_4_elements_with_15_extra_fields-8         20000000           295 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_5_elements_with_15_extra_fields-8         20000000           340 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_0_elements_with_20_extra_fields-8         50000000            77.5 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_1_elements_with_20_extra_fields-8         30000000           163 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_2_elements_with_20_extra_fields-8         20000000           199 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_3_elements_with_20_extra_fields-8         20000000           237 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_4_elements_with_20_extra_fields-8         20000000           287 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_5_elements_with_20_extra_fields-8         20000000           320 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_0_elements_with_25_extra_fields-8         50000000            83.2 ns/op        48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_1_elements_with_25_extra_fields-8         30000000           181 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_2_elements_with_25_extra_fields-8         20000000           222 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_3_elements_with_25_extra_fields-8         20000000           275 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_4_elements_with_25_extra_fields-8         20000000           292 ns/op          48 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_5_elements_with_25_extra_fields-8         20000000           339 ns/op          48 B/op          1 allocs/op

# Request generator

    BenchmarkRequestGeneratePath//a-8                                         10000000           460 ns/op          96 B/op         10 allocs/op
    BenchmarkRequestGeneratePath//a/{{.Supu}}-8                               10000000           522 ns/op         106 B/op         10 allocs/op
    BenchmarkRequestGeneratePath//a?b={{.Tupu}}-8                             10000000           567 ns/op         136 B/op         10 allocs/op
    BenchmarkRequestGeneratePath//a/{{.Supu}}/foo/{{.Foo}}-8                  10000000           615 ns/op         182 B/op         10 allocs/op
    BenchmarkRequestGeneratePath//a/{{.Supu}}/foo/{{.Foo}}/b?c={{.Tupu}}-8    10000000           655 ns/op         236 B/op         10 allocs/op

# Router Handlers

## Gin

    BenchmarkEndpointHandler_ko-8                1000000          5440 ns/op        3026 B/op         31 allocs/op
    BenchmarkEndpointHandler_ok-8                1000000          6456 ns/op        3393 B/op         36 allocs/op
    BenchmarkEndpointHandler_ko_Parallel-8       5000000          1534 ns/op        3028 B/op         31 allocs/op
    BenchmarkEndpointHandler_ok_Parallel-8       5000000          1846 ns/op        3393 B/op         36 allocs/op

## Mux

    BenchmarkEndpointHandler_ko-8                5000000          1815 ns/op        1088 B/op         13 allocs/op
    BenchmarkEndpointHandler_ok-8                5000000          1693 ns/op        1088 B/op         13 allocs/op
    BenchmarkEndpointHandler_ko_Parallel-8      20000000           558 ns/op        1088 B/op         13 allocs/op
    BenchmarkEndpointHandler_ok_Parallel-8      20000000           597 ns/op        1088 B/op         13 allocs/op

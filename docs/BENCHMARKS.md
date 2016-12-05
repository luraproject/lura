Benchmarks
---

Here you'll find some benchmarks of the different components of the KrakenD framework in several scenarios.

# Proxy components

## Proxy middleware stack

    BenchmarkProxyStack_single-8                             1000000         10195 ns/op        2136 B/op         37 allocs/op
    BenchmarkProxyStack_multi/with_1_backends-8              1000000         10289 ns/op        2137 B/op         37 allocs/op
    BenchmarkProxyStack_multi/with_2_backends-8              1000000         18416 ns/op        4512 B/op         82 allocs/op
    BenchmarkProxyStack_multi/with_3_backends-8              1000000         22567 ns/op        6552 B/op        118 allocs/op
    BenchmarkProxyStack_multi/with_4_backends-8               500000         25047 ns/op        8560 B/op        154 allocs/op
    BenchmarkProxyStack_multi/with_5_backends-8               500000         30286 ns/op       10600 B/op        190 allocs/op

## Proxy middlewares

    BenchmarkNewLoadBalancedMiddleware-8                30000000           467 ns/op         328 B/op          6 allocs/op
    BenchmarkNewConcurrentMiddleware_singleNext-8        2000000          9573 ns/op        1376 B/op         21 allocs/op
    BenchmarkNewRequestBuilderMiddleware-8              100000000          125 ns/op         160 B/op          2 allocs/op
    BenchmarkNewMergeDataMiddleware/with_2_parts-8       2000000          7291 ns/op        1536 B/op         22 allocs/op
    BenchmarkNewMergeDataMiddleware/with_3_parts-8       1000000         10499 ns/op        1760 B/op         25 allocs/op
    BenchmarkNewMergeDataMiddleware/with_4_parts-8       1000000         10566 ns/op        1952 B/op         28 allocs/op

# Response manipulation

## Response property whitelisting

    BenchmarkEntityFormatter_whitelistingFilter/with_0_elements_with_0_extra_fields-8           300000000           57.6 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_1_elements_with_0_extra_fields-8           30000000           434 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_2_elements_with_0_extra_fields-8           30000000           485 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_3_elements_with_0_extra_fields-8           30000000           527 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_4_elements_with_0_extra_fields-8           30000000           526 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_0_elements_with_5_extra_fields-8           300000000           56.7 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_1_elements_with_5_extra_fields-8           30000000           526 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_2_elements_with_5_extra_fields-8           30000000           584 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_3_elements_with_5_extra_fields-8           20000000           646 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_4_elements_with_5_extra_fields-8           20000000           647 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_0_elements_with_10_extra_fields-8          300000000           57.2 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_1_elements_with_10_extra_fields-8          20000000           684 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_2_elements_with_10_extra_fields-8          20000000           753 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_3_elements_with_10_extra_fields-8          20000000           809 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_4_elements_with_10_extra_fields-8          20000000           822 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_0_elements_with_15_extra_fields-8          300000000           57.0 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_1_elements_with_15_extra_fields-8          20000000           851 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_2_elements_with_15_extra_fields-8          20000000           949 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_3_elements_with_15_extra_fields-8          20000000          1011 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_4_elements_with_15_extra_fields-8          20000000          1039 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_0_elements_with_20_extra_fields-8          300000000           57.6 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_1_elements_with_20_extra_fields-8          20000000           986 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_2_elements_with_20_extra_fields-8          20000000          1070 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_3_elements_with_20_extra_fields-8          20000000          1145 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_4_elements_with_20_extra_fields-8          20000000          1153 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_0_elements_with_25_extra_fields-8          300000000           57.2 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_1_elements_with_25_extra_fields-8          10000000          1382 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_2_elements_with_25_extra_fields-8          10000000          1504 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_3_elements_with_25_extra_fields-8          10000000          1577 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_whitelistingFilter/with_4_elements_with_25_extra_fields-8          10000000          1603 ns/op         352 B/op          3 allocs/op

## Response property blacklisting

    BenchmarkEntityFormatter_blacklistingFilter/with_0_elements_with_0_extra_fields-8           300000000           56.9 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_1_elements_with_0_extra_fields-8           100000000          131 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_2_elements_with_0_extra_fields-8           100000000          167 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_3_elements_with_0_extra_fields-8           500000000           39.2 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_4_elements_with_0_extra_fields-8           500000000           39.1 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_0_elements_with_5_extra_fields-8           300000000           56.0 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_1_elements_with_5_extra_fields-8           100000000          131 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_2_elements_with_5_extra_fields-8           100000000          167 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_3_elements_with_5_extra_fields-8           100000000          207 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_4_elements_with_5_extra_fields-8           50000000           241 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_0_elements_with_10_extra_fields-8          300000000           56.5 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_1_elements_with_10_extra_fields-8          100000000          130 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_2_elements_with_10_extra_fields-8          100000000          165 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_3_elements_with_10_extra_fields-8          100000000          206 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_4_elements_with_10_extra_fields-8          50000000           244 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_0_elements_with_15_extra_fields-8          300000000           56.8 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_1_elements_with_15_extra_fields-8          100000000          131 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_2_elements_with_15_extra_fields-8          100000000          167 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_3_elements_with_15_extra_fields-8          100000000          206 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_4_elements_with_15_extra_fields-8          50000000           240 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_0_elements_with_20_extra_fields-8          300000000           56.3 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_1_elements_with_20_extra_fields-8          100000000          130 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_2_elements_with_20_extra_fields-8          100000000          167 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_3_elements_with_20_extra_fields-8          100000000          206 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_4_elements_with_20_extra_fields-8          50000000           241 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_0_elements_with_25_extra_fields-8          300000000           56.3 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_1_elements_with_25_extra_fields-8          100000000          137 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_2_elements_with_25_extra_fields-8          100000000          171 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_3_elements_with_25_extra_fields-8          100000000          218 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_blacklistingFilter/with_4_elements_with_25_extra_fields-8          50000000           255 ns/op          16 B/op          1 allocs/op

## Response property groupping

    BenchmarkEntityFormatter_grouping/with_0_elements-8             50000000           285 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_grouping/with_5_elements-8             50000000           305 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_grouping/with_10_elements-8            50000000           303 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_grouping/with_15_elements-8            50000000           303 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_grouping/with_20_elements-8            50000000           305 ns/op         352 B/op          3 allocs/op
    BenchmarkEntityFormatter_grouping/with_25_elements-8            50000000           304 ns/op         352 B/op          3 allocs/op

## Repsonse property mapping

    BenchmarkEntityFormatter_mapping/with_0_elements_with_0_extra_fields-8          300000000           39.9 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_1_elements_with_0_extra_fields-8          300000000           39.7 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_2_elements_with_0_extra_fields-8          300000000           39.5 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_3_elements_with_0_extra_fields-8          300000000           39.3 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_4_elements_with_0_extra_fields-8          300000000           39.3 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_5_elements_with_0_extra_fields-8          300000000           40.0 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_0_elements_with_5_extra_fields-8          200000000           57.4 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_1_elements_with_5_extra_fields-8          100000000          135 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_2_elements_with_5_extra_fields-8          50000000           174 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_3_elements_with_5_extra_fields-8          50000000           200 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_4_elements_with_5_extra_fields-8          50000000           230 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_5_elements_with_5_extra_fields-8          50000000           252 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_0_elements_with_10_extra_fields-8         200000000           57.7 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_1_elements_with_10_extra_fields-8         100000000          137 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_2_elements_with_10_extra_fields-8         50000000           179 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_3_elements_with_10_extra_fields-8         50000000           221 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_4_elements_with_10_extra_fields-8         50000000           271 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_5_elements_with_10_extra_fields-8         30000000           306 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_0_elements_with_15_extra_fields-8         200000000           57.1 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_1_elements_with_15_extra_fields-8         100000000          131 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_2_elements_with_15_extra_fields-8         100000000          167 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_3_elements_with_15_extra_fields-8         50000000           207 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_4_elements_with_15_extra_fields-8         50000000           246 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_5_elements_with_15_extra_fields-8         30000000           283 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_0_elements_with_20_extra_fields-8         200000000           57.4 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_1_elements_with_20_extra_fields-8         100000000          131 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_2_elements_with_20_extra_fields-8         50000000           170 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_3_elements_with_20_extra_fields-8         50000000           210 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_4_elements_with_20_extra_fields-8         50000000           246 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_5_elements_with_20_extra_fields-8         30000000           286 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_0_elements_with_25_extra_fields-8         200000000           57.7 ns/op        16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_1_elements_with_25_extra_fields-8         100000000          131 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_2_elements_with_25_extra_fields-8         50000000           170 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_3_elements_with_25_extra_fields-8         50000000           221 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_4_elements_with_25_extra_fields-8         50000000           261 ns/op          16 B/op          1 allocs/op
    BenchmarkEntityFormatter_mapping/with_5_elements_with_25_extra_fields-8         30000000           300 ns/op          16 B/op          1 allocs/op

# Request generator

    BenchmarkRequestGeneratePath//a-8                                                3000000           427 ns/op          88 B/op          9 allocs/op
    BenchmarkRequestGeneratePath//a/{{.Supu}}-8                                      3000000           475 ns/op          90 B/op          9 allocs/op
    BenchmarkRequestGeneratePath//a?b={{.Tupu}}-8                                    3000000           544 ns/op         120 B/op          9 allocs/op
    BenchmarkRequestGeneratePath//a/{{.Supu}}/foo/{{.Foo}}-8                         3000000           595 ns/op         149 B/op          9 allocs/op
    BenchmarkRequestGeneratePath//a/{{.Supu}}/foo/{{.Foo}}/b?c={{.Tupu}}-8           2000000           757 ns/op         236 B/op         10 allocs/op

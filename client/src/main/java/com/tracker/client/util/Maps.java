package com.tracker.client.util;

import java.util.HashMap;
import java.util.Map;

public class Maps {
  public static <K, V> Map<K, V> of(K k1, V v1) {
    Map<K, V> map = new HashMap<K, V>(1);
    map.put(k1, v1);
    return map;
  }

  public static <K, V> Map<K, V> of(K k1, V v1, K k2, V v2) {
    Map<K, V> map = new HashMap<K, V>(2);
    map.put(k1, v1);
    map.put(k2, v2);
    return map;
  }

  public static <K, V> Map<K, V> of(K k1, V v1, K k2, V v2, K k3, V v3) {
    Map<K, V> map = new HashMap<K, V>(3);
    map.put(k1, v1);
    map.put(k2, v2);
    map.put(k3, v3);
    return map;
  }
}

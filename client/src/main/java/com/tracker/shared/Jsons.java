package com.tracker.shared;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;

import elemental.json.Json;
import elemental.json.JsonArray;
import elemental.json.JsonObject;
import elemental.json.JsonValue;

/**
 * Created by smalex on 13/04/15.
 */
public class Jsons {
  public static JsonObject mapToJsonObject(Map<String, Object> in) {
    final JsonObject object = Json.createObject();
    final Set<String> keys = in.keySet();
    for (String key : keys) {
      final Object inValue = in.get(key);
      if (inValue instanceof String) {
        object.put(key, (String) inValue);
        continue;
      }
      if (inValue instanceof Double) {
        object.put(key, (Double) inValue);
        continue;
      }
      if (inValue instanceof Map) {
        object.put(key, (JsonValue) mapToJsonObject((Map<String, Object>) inValue));
        continue;
      }
      if (inValue instanceof Boolean) {
        object.put(key, (Boolean) inValue);
        continue;
      }
      if (inValue instanceof List) {
        object.put(key, convertArray((List) inValue));
      }
    }
    return object;
  }

  private static JsonArray convertArray(List jsonArray) {
    final JsonArray outArray = Json.createArray();
    for (int i = 0; i < jsonArray.size(); i++) {
      final Object inValue = jsonArray.get(i);
      if (inValue instanceof String) {
        outArray.set(i, (String) inValue);
        continue;
      }
      if (inValue instanceof Double) {
        outArray.set(i, (Double) inValue);
        continue;
      }
      if (inValue instanceof Map) {
        outArray.set(i, (JsonValue) mapToJsonObject((Map<String, Object>) inValue));
        continue;
      }
      if (inValue instanceof Boolean) {
        outArray.set(i, (Boolean) inValue);
        continue;
      }
      if (inValue instanceof List) {
        outArray.set(i, convertArray((List) inValue));
      }
    }
    return outArray;
  }

  public static Map<String, Object> convertJsonObject(JsonObject jsonObject) {
    final Map<String, Object> map = new HashMap<String, Object>();
    for (String key : jsonObject.keys()) {
      map.put(key, getValue(jsonObject.get(key)));
    }
    return map;
  }

  public static Object getValue(JsonValue value) {
    switch (value.getType()) {
      case STRING:
        return value.asString();
      case NUMBER:
        return value.asNumber();
      case BOOLEAN:
        return value.asBoolean();
      case OBJECT:
        return convertJsonObject((JsonObject) value);
      case ARRAY:
        final JsonArray array = (JsonArray) value;
        final List list = new ArrayList();
        for (int i = 0; i < array.length(); i++) {
          list.add(getValue(array.get(i)));
        }
        return list;
      case NULL:
        break;
    }
    return null;
  }
}

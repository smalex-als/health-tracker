package com.tracker.client.util;

import com.google.gwt.core.client.JavaScriptObject;
import com.google.gwt.core.client.JsArray;
import com.google.gwt.core.client.JsArrayString;

import java.util.ArrayList;
import java.util.List;

public class JsArrayUtils {
  public static List<String> toList(JsArrayString array) {
    List<String> result = new ArrayList<String>(array.length());
    for (int i = 0; i < array.length(); i++) {
      result.add(array.get(i));
    }
    return result;
  }

  public static boolean contains(JsArrayString array, String value) {
    for (int i = 0; i < array.length(); i++) {
      if (array.get(i) != null && array.get(i).equals(value)) {
        return true;
      }
    }
    return false;
  }

  public static <T extends JavaScriptObject> List<T> toList(JsArray<T> array) {
    List<T> result = new ArrayList<T>(array.length());
    for (int i = 0; i < array.length(); i++) {
      result.add(array.get(i));
    }
    return result;
  }

  public static <T extends JavaScriptObject> JsArray<T> toArray(List<T> list) {
    JsArray<T> result = JsArray.createArray().cast();
    for (T item : list) {
      result.push(item);
    }
    return result;
  }

  public static JsArrayString toStringArray(List<String> list) {
    final JsArrayString result = JsArrayString.createArray().cast();
    for (String item : list) {
      result.push(item);
    }
    return result;
  }

  public static JsArrayString toStringArray(String ... items) {
    final JsArrayString result = JsArrayString.createArray().cast();
    for (String item : items) {
      result.push(item);
    }
    return result;
  }

  public static <T extends JavaScriptObject> JsArray<T> removeItem(JsArray<T> localItems, int idx) {
    JsArray<T> items = JsArray.createArray().cast();
    for (int i = 0; i < localItems.length(); i++) {
      if (i != idx) {
        items.push(localItems.get(i));
      }
    }
    return items;
  }

  public static JsArrayString parseColumns(String str) {
    JsArrayString result = JsArrayString.createArray().cast();
    for (String pieces : str.split(",")) {
      if (pieces != null && pieces.length() > 0) {
        result.push(pieces);
      }
    }
    return result;
  }
}


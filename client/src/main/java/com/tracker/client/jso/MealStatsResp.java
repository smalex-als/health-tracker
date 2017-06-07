package com.tracker.client.jso;

import com.google.gwt.core.client.JavaScriptObject;
import com.google.gwt.core.client.JsArray;

public class MealStatsResp extends JavaScriptObject {
  public static class Item extends JavaScriptObject {
    protected Item() {}

    public static native Item create() /*-{ 
      return {total: 0, success: false}; 
    }-*/;

    public final native String getDate() /*-{ return this.date; }-*/;

    public final native double getTotal() /*-{ return this.total; }-*/;

    public final native boolean getSuccess() /*-{ return this.success; }-*/;
  }

  protected MealStatsResp() {}
  
  public static native MealStatsResp create() /*-{ 
    return {errors:[], items:[]}; 
  }-*/;

  public final native JsArray<Item> getItems() /*-{ return this.items; }-*/;

  public final native String getDate() /*-{ return this.date; }-*/;

  public final native String getNext() /*-{ return this.next; }-*/;

  public final native String getPrev() /*-{ return this.prev; }-*/;

  public final native JsArray<ErrorJso> getErrors() /*-{ return this.errors; }-*/;
}

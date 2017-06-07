package com.tracker.client.jso;

import com.google.gwt.core.client.JavaScriptObject;

public class MealStatsReq extends JavaScriptObject {
  protected MealStatsReq() {}
  
  public static native MealStatsReq create() /*-{ 
    return {}; 
  }-*/;

  public final native void setDate(String date) /*-{  this.date = date; }-*/;

  public final native String getDate() /*-{ return this.date; }-*/;
}

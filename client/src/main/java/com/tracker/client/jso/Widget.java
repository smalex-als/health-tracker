package com.tracker.client.jso;

import com.google.gwt.core.client.JavaScriptObject;

public class Widget extends JavaScriptObject {
  protected Widget() {}
  
  public static native Widget create() /*-{ return {}; }-*/;

  public final native String getId() /*-{ return this.id; }-*/;

  public final native String getTitle() /*-{ return this.title; }-*/;

  public final native String getType() /*-{ return this.type; }-*/;

  public final native String getDefaultValue() /*-{ return this.defaultValue; }-*/;
}

package com.tracker.client.jso;

import com.google.gwt.core.client.JavaScriptObject;

public class ErrorJso extends JavaScriptObject {
  protected ErrorJso() {}

  public static native Error create() /*-{ return {code: 0}; }-*/;

  public final native String getField() /*-{ return this.field; }-*/;

  public final native String getMessage() /*-{ return this.message; }-*/;

  public final native int getCode() /*-{ return this.code; }-*/;
}

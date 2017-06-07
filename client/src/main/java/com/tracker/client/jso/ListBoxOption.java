package com.tracker.client.jso;

import com.google.gwt.core.client.JavaScriptObject;

public class ListBoxOption extends JavaScriptObject {
  protected ListBoxOption() {}
  
  public static native ListBoxOption create() /*-{ return {}; }-*/;

  public final native String getValue() /*-{ return this.value; }-*/;

  public final native String getName() /*-{ return this.name; }-*/;
}

package com.tracker.client.jso;

import com.google.gwt.core.client.JavaScriptObject;
import com.google.gwt.core.client.JsArray;

public class EditView extends JavaScriptObject {
  protected EditView() {}
  
  public static native EditView create() /*-{ 
    return {widgets:[]}; 
  }-*/;

  public final native JsArray<Widget> getWidgets() /*-{ return this.widgets; }-*/;
}

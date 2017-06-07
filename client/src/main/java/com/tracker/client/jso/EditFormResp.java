package com.tracker.client.jso;

import com.google.gwt.core.client.JavaScriptObject;
import com.google.gwt.core.client.JsArray;

public class EditFormResp extends JavaScriptObject {
  protected EditFormResp() {}
  
  public static native EditFormResp create() /*-{ return {}; }-*/;

  public final native EditView getEditView() /*-{ return this.editView; }-*/;

  public final native JsArray<ListBoxOption> getReferences(String name) /*-{
    return this.references[name];
  }-*/;
}

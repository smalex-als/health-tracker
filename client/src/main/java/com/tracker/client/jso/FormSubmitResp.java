package com.tracker.client.jso;

import com.google.gwt.core.client.JavaScriptObject;
import com.google.gwt.core.client.JsArray;

public class FormSubmitResp extends JavaScriptObject {
  protected FormSubmitResp() {}
  
  public static native FormSubmitResp create() /*-{ 
    return {errors: []}; 
  }-*/;

  public final native String getStatus() /*-{ return this.status; }-*/;

  public final native User getUser()  /*-{ return this.user; }-*/;

  public final native JsArray<ErrorJso> getErrors() /*-{ return this.errors; }-*/;
}

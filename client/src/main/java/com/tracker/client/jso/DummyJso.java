package com.tracker.client.jso;

import com.google.gwt.core.client.JsArray;

public class DummyJso {
  protected DummyJso() {}
  
  public static native DummyJso create() /*-{ 
    return {errors:[]}; 
  }-*/;

  public final native JsArray<ErrorJso> getErrors() /*-{ return this.errors; }-*/;
}

package com.tracker.client.jso;

import com.google.gwt.core.client.JavaScriptObject;

public class SignInReq extends JavaScriptObject {
  protected SignInReq() {}
  
  public static native SignInReq create() /*-{ return {errors: []}; }-*/;

  public final native String getUsername() /*-{ return this.username; }-*/;

  public final native void setUsername(String username) /*-{  this.username = username; }-*/;

  public final native String getPassword() /*-{ return this.password; }-*/;

  public final native void setPassword(String password) /*-{  this.password = password; }-*/;
}

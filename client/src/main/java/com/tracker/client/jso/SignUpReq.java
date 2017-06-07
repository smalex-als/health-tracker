package com.tracker.client.jso;

import com.google.gwt.core.client.JavaScriptObject;

public class SignUpReq extends JavaScriptObject {
  protected SignUpReq() {}
  
  public static native SignUpReq create() /*-{ return {errors: []}; }-*/;

  public final native String getEmail() /*-{ return this.email; }-*/;

  public final native void setEmail(String email) /*-{  this.email = email; }-*/;

  public final native String getUsername() /*-{ return this.username; }-*/;

  public final native void setUsername(String username) /*-{  this.username = username; }-*/;

  public final native String getPassword() /*-{ return this.password; }-*/;

  public final native void setPassword(String password) /*-{  this.password = password; }-*/;
}

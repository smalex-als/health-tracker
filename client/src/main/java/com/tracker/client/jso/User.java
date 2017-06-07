package com.tracker.client.jso;

import com.google.gwt.core.client.JavaScriptObject;

public class User extends JavaScriptObject {
  protected User() {}
  
  public static native User create() /*-{ 
    return {enabled: false, role: 0}; 
  }-*/;

  public final native String getId() /*-{ return this.id; }-*/;

  public final native void setId(String id) /*-{  this.id = id; }-*/;

  public final native boolean getEnabled() /*-{ return this.enabled; }-*/;

  public final native void setEnabled(boolean enabled) /*-{  this.enabled = enabled; }-*/;

  public final native void setEmailConfirmed(boolean emailConfirmed) /*-{ this.emailConfirmed = emailConfirmed; }-*/;

  public final native boolean getEmailConfirmed() /*-{ return this.emailConfirmed; }-*/;

  public final native String getEmail() /*-{ return this.email; }-*/;

  public final native void setEmail(String email) /*-{  this.email = email; }-*/;

  public final native String getUsername() /*-{ return this.username; }-*/;

  public final native void setUsername(String username) /*-{  this.username = username; }-*/;

  public final native String getNewPassword() /*-{ return this.new_password; }-*/;

  public final native void setNewPassword(String newPassword) /*-{  this.new_password = newPassword; }-*/;

  public final native int getRole() /*-{ return this.role; }-*/;

  public final native void setRole(int role) /*-{  this.role = role; }-*/;

  public final native String getRoleId() /*-{ return this.roleId; }-*/;

  public final native String getCreated() /*-{ return this.created; }-*/;

  public final native void setCreated(String created) /*-{  this.created = created; }-*/;
}

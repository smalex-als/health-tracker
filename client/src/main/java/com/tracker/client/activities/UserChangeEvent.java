package com.tracker.client.activities;

import com.tracker.client.jso.User;
import com.google.web.bindery.event.shared.Event;
import com.google.web.bindery.event.shared.EventBus;
import com.google.web.bindery.event.shared.HandlerRegistration;

public class UserChangeEvent extends Event<UserChangeEvent.Handler> {
  private static final Type<UserChangeEvent.Handler> TYPE = new Type<UserChangeEvent.Handler>();

  private User user;

  public static void fire(EventBus eventBus, User user, String sourceName) {
    eventBus.fireEventFromSource(new UserChangeEvent(user), sourceName);
  }

  public static HandlerRegistration register(EventBus eventBus, String sourceName, Handler handler) {
    return eventBus.addHandlerToSource(TYPE, sourceName, handler);
  }

  public UserChangeEvent(User user) {
    this.user = user;
  }

  @Override
  public Type<Handler> getAssociatedType() {
    return TYPE;
  }

  @Override
  protected void dispatch(Handler handler) {
    handler.onUserChange(this);
  }

  public User getUser() {
    return user;
  }

  public interface Handler {
    void onUserChange(UserChangeEvent event);
  }
}

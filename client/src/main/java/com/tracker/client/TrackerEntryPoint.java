package com.tracker.client;

import java.util.logging.Logger;

import com.tracker.client.activities.BaseActivity.MyActivityWidget;
import com.tracker.client.activities.PageWrapCompontent;
import com.tracker.client.controls.Component;
import com.tracker.client.jso.User;
import com.tracker.client.place.BasePlace;
import com.tracker.client.util.StringUtils;
import com.google.gwt.activity.shared.ActivityManager;
import com.google.gwt.core.client.EntryPoint;
import com.google.gwt.place.shared.PlaceController;
import com.google.gwt.place.shared.PlaceHistoryHandler;
import com.google.gwt.user.client.ui.AcceptsOneWidget;
import com.google.gwt.user.client.ui.IsWidget;
import com.google.web.bindery.event.shared.EventBus;

import elemental.client.Browser;
import elemental.dom.Document;
import elemental.dom.Element;

public class TrackerEntryPoint implements EntryPoint, AcceptsOneWidget {
  private static final Logger log = Logger.getLogger(TrackerEntryPoint.class.getName());
  private final AppFactory factory = new AppFactory();
  private PageWrapCompontent pageWrapCompontent;
  private Component currentActivity;

  @Override
  public void onModuleLoad() {
    // GWT.setUncaughtExceptionHandler(new GWT.UncaughtExceptionHandler() {
    //   @Override
    //   public void onUncaughtException(Throwable e) {
    //     Window.alert("catch: " + e.toString());
    //     log.log(Level.SEVERE, "catch: " + e.toString(), e);
    //   }
    // });
    //

    Document doc = Browser.getDocument();
    Element progressEl = doc.createDivElement();
    progressEl.setId("progress");
    progressEl.setInnerHTML("<dt/><dd/>");
    // progressEl.getStyle().setOpacity(0);
    progressEl.setClassName("start");
    doc.getBody().appendChild(progressEl);

    User user = getUserJson();
    if (StringUtils.hasText(user.getUsername())) {
      factory.setCurrentUser(user);
    }

    final ActivityManager activityManager = factory.getActivityManager();
    activityManager.setDisplay(this);

    PlaceHistoryHandler placeHistoryHandler = factory.getPlaceHistoryHandler();
    PlaceController placeController = factory.getPlaceController();
    EventBus eventBus = factory.getEventBus();

    BasePlace nextPlace = null;
    if (!StringUtils.hasText(user.getUsername())) {
      nextPlace = factory.getDefaultPlace();
    } else if (user.getEmailConfirmed()) {
      nextPlace = factory.getNextPlace();
    } else {
      nextPlace = factory.getConfirmEmailPlace();
    }
    placeHistoryHandler.register(placeController, eventBus, nextPlace);
    placeHistoryHandler.handleCurrentHistory();
  }

  private static native User getUserJson() /*-{
    return $wnd.User != undefined ? $wnd.User : {};
  }-*/;

  @Override
  public void setWidget(IsWidget w) {
    removeCurrentActivity();
    if (w instanceof MyActivityWidget) {
      currentActivity = ((MyActivityWidget)w).getComponent();

      if (pageWrapCompontent == null) {
        pageWrapCompontent = new PageWrapCompontent(factory);
        Document doc = Browser.getDocument();
        Element content = doc.getElementById("content");
        pageWrapCompontent.render(content);
      }
      pageWrapCompontent.addChild(currentActivity, true);
    }
  }

  private void removeCurrentActivity() {
    if (currentActivity != null) {
      pageWrapCompontent.removeChild(currentActivity);
      currentActivity = null;
    }
  }
}

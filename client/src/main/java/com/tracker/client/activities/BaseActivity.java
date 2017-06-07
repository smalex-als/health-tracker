package com.tracker.client.activities;

import java.util.logging.Logger;

import com.tracker.client.AppFactory;
import com.tracker.client.controls.Component;
import com.tracker.client.place.BasePlace;
import com.google.gwt.activity.shared.Activity;
import com.google.gwt.event.shared.EventBus;
import com.google.gwt.event.shared.HandlerRegistration;
import com.google.gwt.user.client.Event;
import com.google.gwt.user.client.Event.NativePreviewEvent;
import com.google.gwt.user.client.ui.AcceptsOneWidget;
import com.google.gwt.user.client.ui.IsWidget;
import com.google.gwt.user.client.ui.Widget;

import elemental.dom.Element;

public abstract class BaseActivity extends Component implements Activity {
  private final Logger log = Logger.getLogger(getClass().getName());
  // private final GlobalStyleSwitcher globalStyleSwitcher;
  private HandlerRegistration nativePreviewHandlerRegistration;
  protected AppFactory factory;

  public BaseActivity(AppFactory factory) {
    this.factory = factory;
    // globalStyleSwitcher = new GlobalStyleSwitcher(factory);
  }

  public String mayStop() {
    return null;
  }

  public void onCancel() {
  }

  public void onStop() {
  }

  public static class MyActivityWidget extends Widget implements IsWidget {
    private final Component component;
    public MyActivityWidget(Component component) {
      this.component = component;
      setElement((com.google.gwt.dom.client.Element) component.getElement());
    }

    public Component getComponent() {
      return component;
    }
  }


  public static interface StartCallback {
    void start();
  }

  public void start(AcceptsOneWidget panel, EventBus eventBus) {
    StartCallback startCallback = new StartCallback() {
      @Override
      public void start() {
        log.info("start");
        panel.setWidget(new MyActivityWidget(BaseActivity.this));
      }
    };
    start(startCallback);
  }

  public void start(StartCallback callback) {
    callback.start();
  }

  public void updateForPlace(BasePlace basePlace) {
    log.info("updateForPlace " + basePlace);
  }

  @Override
  public void decorateInternal(final Element element) {
    super.decorateInternal(element);
  }

  @Override
  public void enterDocument() {
    log.info("enterDocument");
    super.enterDocument();
    // globalStyleSwitcher.enterDocument();
    startHandlingKeys();
  }

  @Override
  public void exitDocument() {
    log.info("exitDocument");
    stopHandlingKeys();
    super.exitDocument();
  }

  protected boolean onPreviewNativeEvent(NativePreviewEvent event) {
    return false;
  }

  private void startHandlingKeys() {
    stopHandlingKeys();
    // log.info("startHandlingKeys");
    nativePreviewHandlerRegistration = Event.addNativePreviewHandler(new Event.NativePreviewHandler() {
      @Override
      public void onPreviewNativeEvent(Event.NativePreviewEvent event) {
        // log.info("hit key " + nativeEvent);
        if (!BaseActivity.this.onPreviewNativeEvent(event)) {
//           globalStyleSwitcher.onPreviewNativeEvent(event);
        }
      }
    });
  }

  private void stopHandlingKeys() {
    // stop previewing page events
    if (nativePreviewHandlerRegistration != null) {
      // log.info("stopHandlingKeys");
      nativePreviewHandlerRegistration.removeHandler();
      nativePreviewHandlerRegistration = null;
    }
  }
}


package com.tracker.client.activities;

import java.util.logging.Logger;

import com.tracker.client.AppFactory;
import com.tracker.client.controls.Component;
import com.tracker.client.jso.User;
import com.tracker.client.rpc.ContentRpcService;

import elemental.dom.Element;

public class PageWrapCompontent extends Component implements UserChangeEvent.Handler {
  private static final Logger log = Logger.getLogger(PageWrapCompontent.class.getName());
  private ContentRpcService rpc;
  private AppFactory factory;
  private final NavBarComponent navBar;

  public PageWrapCompontent(AppFactory factory) {
    this.factory = factory;
    rpc = factory.getRpcService();
    navBar = new NavBarComponent(factory);
  }

  @Override
  public void decorateInternal(final Element element) {
    super.decorateInternal(element);
  }

  @Override
  public void enterDocument() {
    super.enterDocument();
    UserChangeEvent.register(factory.getEventBus(), "", this);
    updateNavBar(factory.getCurrentUser());
  }

  @Override
  public void onUserChange(UserChangeEvent event) {
    updateNavBar(event.getUser());
  }

  private void updateNavBar(User user) {
    if (user != null) {
      if (!navBar.isInDocument()) {
        addChildAt(navBar, 0, true);
      }
      navBar.updateView(user);
    } else {
      removeChild(navBar);
    }
  }
}

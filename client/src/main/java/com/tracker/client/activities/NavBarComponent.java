package com.tracker.client.activities;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.logging.Logger;

import com.tracker.client.AppFactory;
import com.tracker.client.controls.Component;
import com.tracker.client.jso.User;
import com.tracker.client.rpc.ContentRpcService;
import com.tracker.client.rpc.ContentRpcService.AsyncCallback;
import com.tracker.client.util.DomUtils;
import com.tracker.client.util.Maps;
import com.tracker.client.util.NumberUtils;
import com.tracker.client.util.StyleUtils;
import com.tracker.shared.NavBarTemplate;

import elemental.dom.Element;
import elemental.events.Event;
import elemental.events.EventListener;
import elemental.json.JsonObject;

public class NavBarComponent extends Component {
  private static final Logger log = Logger.getLogger(NavBarComponent.class.getName());
  private NavBarTemplate template = new NavBarTemplate();
  private Element usernameEl;
  private Element btnLogout;
  private AppFactory factory;
  private ContentRpcService rpc;

  public NavBarComponent(AppFactory factory) {
    this.factory = factory;
    rpc = factory.getRpcService();
  }

  @Override
  protected void createDom() {
    Map<String, Object> map = new HashMap<>();
    List<Map<String, Object>> menuItems = new ArrayList<>();
    menuItems.add(Maps.of("name", "Meals", "href", "#l:meals"));
    menuItems.add(Maps.of("name", "Stats", "href", "#l:meals_stats"));
    menuItems.add(Maps.of("name", "Users", "href", "#l:users", "roleId", "2"));
    menuItems.add(Maps.of("name", "Settings", "href", "#l:settings"));
    map.put("items", menuItems);

    String body = template.toString(template.renderNavBar(map));
    decorateInternal((Element) DomUtils.htmlToDocumentFragment_(doc, body));
  }

  @Override
  public void decorateInternal(final Element element) {
    super.decorateInternal(element);

    btnLogout = getElementByClassName("btn-logout");
    usernameEl = getElementByClassName("user-username");
  }

  @Override
  public void enterDocument() {
    super.enterDocument();
    User user = factory.getCurrentUser();
    usernameEl.setInnerText(user.getUsername());
    addHandlerRegistration(btnLogout.addEventListener(Event.CLICK, new EventListener() {
      @Override
      public void handleEvent(Event evt) {
        evt.stopPropagation();
        evt.preventDefault();
        handleLogout();
      }
    }, false));
  }

  public void updateView(User user) {
    usernameEl.setInnerText(user.getUsername());
    Element[] els = getElementsByClassName("nav-item");
    if (user.getEmailConfirmed()) {
      for (int i = 0; i < els.length; i++) {
        StyleUtils.showElement(els[i], null);
      }
      // TODO little hack need to be fixed
      if (NumberUtils.toInt(user.getRoleId()) > 1) {
        StyleUtils.showElement(els[els.length - 3], null);
      } else {
        StyleUtils.showElement(els[els.length - 3], "none");
      }
    } else {
      for (int i = 0; i < els.length - 1; i++) {
        StyleUtils.showElement(els[i], "none");
      }
    }
  }

  private void handleLogout() {
    rpc.request("/v1/users/signout/", null, new AsyncCallback<JsonObject>() {
      @Override
      public void onSuccess(JsonObject resp) {
        factory.setCurrentUser(null);
        factory.getPlaceController().goTo(factory.getDefaultPlace());
      }
    });
  }
}

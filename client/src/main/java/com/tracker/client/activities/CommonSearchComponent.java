package com.tracker.client.activities;

import java.util.HashMap;
import java.util.Map;

import com.tracker.client.controls.Component;
import com.tracker.client.util.DomUtils;
import com.tracker.client.util.StringUtils;
import com.tracker.shared.TableTemplate;

import elemental.dom.Element;
import elemental.events.Event;
import elemental.events.EventListener;
import elemental.html.FormElement;
import elemental.html.InputElement;
import elemental.json.Json;
import elemental.json.JsonObject;

public class CommonSearchComponent extends Component implements SearchComponent {
  private InputElement searchInputEl;
  private FormElement searchFormEl;
  private Presenter presenter;

  @Override
  protected void createDom() {
    Map<String, Object> map = new HashMap<>();
    TableTemplate template = new TableTemplate();
    String body = template.toString(template.renderSearch(map));
    decorateInternal((Element) DomUtils.htmlToDocumentFragment_(doc, body));
  }

  @Override
  public void decorateInternal(final Element element) {
    super.decorateInternal(element);
    searchInputEl = (InputElement) getElementByClassName("search-input");
    searchFormEl = (FormElement) getElementByClassName("search-form");
  }

  @Override
  public void enterDocument() {
    super.enterDocument();

    addHandlerRegistration(searchFormEl.addEventListener(Event.SUBMIT, new EventListener() {
      @Override
      public void handleEvent(Event evt) {
        evt.stopPropagation();
        evt.preventDefault();
        if (presenter != null) {
          presenter.clickSearch();
        }
      }
    }, false));
  }

  @Override
  public JsonObject updateModel() {
    JsonObject map = Json.createObject();
    map.put("query", StringUtils.trimToEmpty(searchInputEl.getValue()));
    return map;
  }

  @Override
  public void updateView(JsonObject in) {
    searchInputEl.setValue(StringUtils.trimToEmpty(in.getString("query")));
  }

  @Override
  public void setPresenter(Presenter presenter) {
    this.presenter = presenter;
  }

  @Override
  public Component getComponent() {
    return this;
  }
}

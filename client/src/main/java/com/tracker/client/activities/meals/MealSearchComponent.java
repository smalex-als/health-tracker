package com.tracker.client.activities.meals;

import java.util.Date;
import java.util.HashMap;
import java.util.Map;

import com.tracker.client.activities.SearchComponent;
import com.tracker.client.controls.Component;
import com.tracker.client.util.DomUtils;
import com.tracker.client.util.NumberUtils;
import com.tracker.client.util.StringUtils;
import com.tracker.shared.TableTemplate;
import com.google.gwt.i18n.client.DateTimeFormat;

import elemental.dom.Element;
import elemental.events.Event;
import elemental.events.EventListener;
import elemental.html.FormElement;
import elemental.html.InputElement;
import elemental.html.OptionElement;
import elemental.html.SelectElement;
import elemental.json.Json;
import elemental.json.JsonObject;

public class MealSearchComponent extends Component implements SearchComponent {
  private static DateTimeFormat dateFormat = DateTimeFormat.getFormat("yyyy-MM-dd");
  private Presenter presenter;
  private FormElement formEl;
  private InputElement dateEl;
  private SelectElement searchDaysEl;
  private InputElement searchInputEl;
  private InputElement fromTime;
  private InputElement toTime;

  @Override
  protected void createDom() {
    Map<String, Object> map = new HashMap<>();
    map.put("dateSearch", true);
    TableTemplate template = new TableTemplate();
    String body = template.toString(template.renderSearch(map));
    decorateInternal((Element) DomUtils.htmlToDocumentFragment_(doc, body));
  }

  @Override
  public void decorateInternal(final Element element) {
    super.decorateInternal(element);
    
    formEl = (FormElement) getElementByClassName("search-form");
    searchInputEl = (InputElement) getElementByClassName("search-input");
    dateEl = (InputElement) getElementByClassName("search-bydate");
    searchDaysEl = (SelectElement) getElementByClassName("search-with-days");
    fromTime = (InputElement) getElementByClassName("search-fromtime");
    toTime = (InputElement) getElementByClassName("search-totime");
  }

  @Override
  public void enterDocument() {
    super.enterDocument();

    addHandlerRegistration(formEl.addEventListener(Event.SUBMIT, new EventListener() {
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
    JsonObject model = Json.createObject();
    model.put("query", StringUtils.trimToEmpty(searchInputEl.getValue()));
    model.put("date", StringUtils.trimToEmpty(dateEl.getValue()));
    model.put("days", getSelectedValue(searchDaysEl));
    model.put("from", getValueTime(fromTime.getValue()));
    model.put("to", getValueTime(toTime.getValue()));
    return model;
  }

  @Override
  public void updateView(JsonObject in) {
    searchInputEl.setValue(StringUtils.trimToEmpty(in.getString("query")));

    int days = NumberUtils.toInt(in.getString("days"));
    days = Math.max(0, Math.min(days, 60));
    if (days == 0) {
      days = 3;
    }

    setSelectedValue(searchDaysEl, String.valueOf(days));
    searchInputEl.setValue(StringUtils.trimToEmpty(in.getString("query")));
    String date = StringUtils.trimToEmpty(in.getString("date"));
    if (date.length() > 0) {
      dateEl.setValue(date);
    } else {
      dateEl.setValue(dateToString(new Date()));
    }
    fromTime.setValue(setValueTime(in.getString("from")));
    toTime.setValue(setValueTime(in.getString("to")));
  }

  @Override
  public void setPresenter(Presenter presenter) {
    this.presenter = presenter;
  }

  @Override
  public Component getComponent() {
    return this;
  }

  private String dateToString(Date date) {
    return dateFormat.format(date);
  }

  private String getSelectedValue(SelectElement input) {
    OptionElement el = (OptionElement) input.getOptions().item(input.getSelectedIndex());
    return el.getValue();
  }

  private void setSelectedValue(SelectElement selectEl, String value) {
    for (int i = 0; i < selectEl.getOptions().length(); i++) {
      OptionElement el = (OptionElement) selectEl.getOptions().item(i);
      el.setSelected(el.getValue().equals(value));
    }
  }

  private String getValueTime(String val) {
    if (val.length() > 3) {
      return val.substring(0, 2) + val.substring(3);
    }
    return "";
  }

  private String setValueTime(String val) {
    if (val == null || val == "") {
      return "";
    }
    if (val.length() < 4) {
      val = "0000".substring(val.length()) + val;
    }
    return val.substring(0, 2)  + ":" + val.substring(2);
  }
}

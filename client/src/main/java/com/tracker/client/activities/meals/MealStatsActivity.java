package com.tracker.client.activities.meals;

import java.util.Date;
import java.util.HashMap;
import java.util.Map;
import java.util.logging.Logger;

import com.tracker.client.AppFactory;
import com.tracker.client.activities.BaseActivity;
import com.tracker.client.jso.MealStatsReq;
import com.tracker.client.jso.MealStatsResp;
import com.tracker.client.place.BasePlace;
import com.tracker.client.rpc.ContentRpcService;
import com.tracker.client.util.DomUtils;
import com.tracker.client.util.FormUtils;
import com.tracker.client.util.StringUtils;
import com.tracker.shared.Jsons;
import com.tracker.shared.MealsTemplate;
import com.tracker.shared.TableTemplate;
import com.google.gwt.i18n.client.DateTimeFormat;

import elemental.dom.Element;
import elemental.events.Event;
import elemental.events.EventListener;
import elemental.json.JsonObject;

public class MealStatsActivity extends BaseActivity {
  private static final Logger log = Logger.getLogger(MealStatsActivity.class.getName());
  private boolean running;
  private BasePlace place;
  private ContentRpcService rpc;
  private Element pageNext;
  private Element pageCur;
  private Element pagePrev;
  private Element tableEl;
  private String nextDate;
  private String prevDate;
  private Element totalEl;
  private FormUtils formUtils = new FormUtils(this);
  private MealsTemplate template = new MealsTemplate();
  private static DateTimeFormat dateFormat = DateTimeFormat.getFormat("yyyy-MM-dd");

  public MealStatsActivity(AppFactory factory) {
    super(factory);
    this.rpc = factory.getRpcService();
  }

  @Override
  protected void createDom() {
    Map<String, Object> map = new HashMap<>();
    TableTemplate template = new TableTemplate();
    String body = template.toString(template.renderPagination(map));
    decorateInternal((Element) DomUtils.htmlToDocumentFragment_(doc, body));
  }

  @Override
  public void decorateInternal(final Element element) {
    super.decorateInternal(element);
    pagePrev = getElementByClassNameRequired("page-prev");
    pageNext = getElementByClassNameRequired("page-next");
    pageCur = getElementByClassNameRequired("page-current");
    tableEl = getElementByClassNameRequired("meals-stats");
  }

  public void updateForPlace(BasePlace place) {
    this.place = place;
    if (running) {
      // reload for new place
      start(null);
    }
  }

  @Override
  public void enterDocument() {
    super.enterDocument();
    running = true;

    addHandlerRegistration(pageNext.addEventListener(Event.CLICK, new EventListener() {
      @Override
      public void handleEvent(Event evt) {
        handleOpen(nextDate);
        evt.stopPropagation();
        evt.preventDefault();
      }
    }, false));
    addHandlerRegistration(pagePrev.addEventListener(Event.CLICK, new EventListener() {
      @Override
      public void handleEvent(Event evt) {
        handleOpen(prevDate);
        evt.stopPropagation();
        evt.preventDefault();
      }
    }, false));
  }

  @Override
  public void exitDocument() {
    running = false;
    super.exitDocument();
  }

  @Override
  public void start(StartCallback callback) {
    if (getElement() == null) {
      createDom();
    }

    String date = place.getParam("date");
    if (!StringUtils.hasText(date)) {
      date = dateFormat.format(new Date());
    }
    MealStatsReq in = MealStatsReq.create();
    in.setDate(date);
    rpc.request("GET", "/v1/meals-stats/", in, 
        (MealStatsResp out) -> handleResponse(out, callback));
  }

  private void handleOpen(String date) {
    factory.getPlaceController().goTo(
        BasePlace.newBuilder()
        .viewId(place.getViewId()).param("date", date).build());
  }

  private void handleResponse(MealStatsResp resp, StartCallback callback) {
    if (!formUtils.updateViewErrors(resp.getErrors()) ) {
      nextDate = resp.getNext();
      prevDate = resp.getPrev();
      ((Element)pagePrev.getFirstChild()).setInnerText("<< " + prevDate);
      ((Element)pageCur.getFirstChild()).setInnerText(resp.getDate());
      ((Element)pageNext.getFirstChild()).setInnerText(nextDate + " >>");
      //    totalEl.setInnerText("Total: " + formatDouble(resp.getTotal()));
      Map<String, Object> map = Jsons.convertJsonObject((JsonObject) resp);
      log.info("map = " + map);
      tableEl.setInnerHTML(template.toString(template.renderStats(map)));
      if (callback != null) {
        callback.start();
      }
    }
  }

  private String formatDouble(double val) {
    String str = String.valueOf(val);
    int index = str.indexOf('.');
    if (index >= 0) {
      return str.substring(0, Math.min(str.length(), index + 3));
    }
    return str;
  }
}

